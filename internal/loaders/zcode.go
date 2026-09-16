package loaders

import (
	"database/sql"
	"fmt"

	"github.com/jadmadi/thermal/internal/thermal"

	_ "modernc.org/sqlite"
)

// LoadZCodeData reads the ZCode CLI SQLite DB (~/.zcode/cli/db/db.sqlite).
// ZCode is an OpenCode-family fork: it keeps the session/message/part schema
// but adds per-request telemetry tables. model_usage is the primary source,
// with pre-aggregated token columns (input, output, reasoning, cache
// creation/read) plus provider, model, and agent on every completed request.
// turn_usage supplies user-visible turn counts for daily rows. The DB records
// no cost figures, so Cost stays 0.
func LoadZCodeData(dbPath string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=30000000000")
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	defer db.Close()
	_, _ = db.Exec("PRAGMA cache_size = -64000; PRAGMA mmap_size = 30000000000;")

	var hasModelUsage bool
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='model_usage'`).Scan(&hasModelUsage); err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	if !hasModelUsage {
		return thermal.Summary{}, nil, nil, fmt.Errorf("zcode: model_usage table not found in %s", dbPath)
	}

	var summary thermal.Summary
	// ZCode follows the provider convention where input_tokens already
	// contains cache reads and reasoning is part of output_tokens, so both are
	// reported without the nested part. Otherwise the type columns double
	// count against LifetimeTokens.
	err = db.QueryRow(`
		SELECT
			COUNT(DISTINCT session_id),
			COALESCE(SUM(computed_total_tokens), 0),
			COALESCE(SUM(MAX(input_tokens - cache_read_input_tokens - cache_creation_input_tokens, 0)), 0),
			COALESCE(SUM(MAX(output_tokens - reasoning_tokens, 0)), 0),
			COALESCE(SUM(reasoning_tokens), 0),
			COALESCE(SUM(cache_creation_input_tokens + cache_read_input_tokens), 0)
		FROM model_usage
		WHERE status = 'completed'
	`).Scan(&summary.Sessions, &summary.LifetimeTokens, &summary.InputTokens,
		&summary.OutputTokens, &summary.ReasoningTokens, &summary.CacheTokens)
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}

	// Longest session duration from the session table when present.
	_ = db.QueryRow(`SELECT COALESCE(MAX(time_updated - time_created), 0) FROM session`).
		Scan(&summary.LongestSessionMs)

	// Code-change analytics from session summary columns (NULL on current
	// schema, COALESCE keeps them 0 until ZCode populates them).
	_ = db.QueryRow(`
		SELECT COALESCE(SUM(summary_additions), 0), COALESCE(SUM(summary_deletions), 0), COALESCE(SUM(summary_files), 0)
		FROM session WHERE summary_additions IS NOT NULL
	`).Scan(&summary.LinesAdded, &summary.LinesDeleted, &summary.FilesTouched)

	// Model and agent distribution from completed requests.
	if modelRows, err := db.Query(`SELECT model_id, COUNT(*) FROM model_usage WHERE status = 'completed' AND model_id != '' GROUP BY model_id`); err == nil {
		summary.ModelBreakdown = make(map[string]int64)
		for modelRows.Next() {
			var model string
			var n int64
			modelRows.Scan(&model, &n)
			summary.ModelBreakdown[model] = n
		}
		modelRows.Close()
	}
	if agentRows, err := db.Query(`SELECT agent, COUNT(*) FROM model_usage WHERE status = 'completed' AND agent != '' GROUP BY agent`); err == nil {
		summary.AgentBreakdown = make(map[string]int)
		for agentRows.Next() {
			var agent string
			var n int
			agentRows.Scan(&agent, &n)
			summary.AgentBreakdown[agent] = n
		}
		agentRows.Close()
	}

	// Daily tokens from completed model requests, grouped by day and model so
	// reports can break usage down per model. Turns are counted once per day:
	// completed turn_usage rows when present (one per user-visible turn), else
	// completed request counts.
	var hasTurns bool
	_ = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='turn_usage'`).Scan(&hasTurns)

	turnTable := "model_usage"
	if hasTurns {
		turnTable = "turn_usage"
	}
	// Turns per day and project. turn_usage has no project column, so the
	// session row supplies it when the schema carries a directory.
	turnJoin, turnProject := "", "''"
	if hasColumn(db, "session", "directory") {
		turnProject = "COALESCE(NULLIF(s.directory, ''), '')"
		turnJoin = " LEFT JOIN session s ON s.id = t.session_id"
	}
	turnsByDay := make(map[string]int)
	turnsByKey := make(map[projectDayKey]int)
	if turnRows, err := db.Query(`
		SELECT date(t.started_at / 1000, 'unixepoch', 'localtime') AS day,
			` + turnProject + ` AS project,
			COUNT(*)
		FROM ` + turnTable + ` t` + turnJoin + `
		WHERE t.status = 'completed'
		GROUP BY day, project
	`); err == nil {
		for turnRows.Next() {
			var day, project string
			var n int
			if turnRows.Scan(&day, &project, &n) == nil {
				turnsByDay[day] += n
				if key := thermal.ProjectKey(project); key != "" {
					turnsByKey[projectDayKey{day, key}] += n
				}
			}
		}
		turnRows.Close()
	}

	modelJoin, modelProject := "", "''"
	if hasColumn(db, "session", "directory") {
		modelProject = "COALESCE(NULLIF(s.directory, ''), '')"
		modelJoin = " LEFT JOIN session s ON s.id = m.session_id"
	}
	rows, err := db.Query(`
		SELECT
			date(m.started_at / 1000, 'unixepoch', 'localtime') AS day,
			` + modelProject + ` AS project,
			COALESCE(m.model_id, '') AS model,
			COALESCE(SUM(MAX(m.input_tokens - m.cache_read_input_tokens - m.cache_creation_input_tokens, 0)), 0),
			COALESCE(SUM(MAX(m.output_tokens - m.reasoning_tokens, 0)), 0),
			COALESCE(SUM(m.reasoning_tokens), 0),
			COALESCE(SUM(m.cache_read_input_tokens), 0),
			COALESCE(SUM(m.cache_creation_input_tokens), 0),
			0.0,
			0,
			COALESCE(SUM(m.computed_total_tokens), 0)
		FROM model_usage m` + modelJoin + `
		WHERE m.status = 'completed'
		GROUP BY day, project, model
		ORDER BY day
	`)
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	defer rows.Close()

	daily, projects, err := foldDayModelProjectRows(rows)
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	for i := range daily {
		if n, ok := turnsByDay[daily[i].Day]; ok {
			daily[i].Turns = n
		}
	}
	for i := range projects {
		if n, ok := turnsByKey[projectDayKey{projects[i].Day, projects[i].Project}]; ok {
			projects[i].Turns = n
		}
	}

	return summary, daily, projects, nil
}
