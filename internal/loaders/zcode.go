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
func LoadZCodeData(dbPath string) (thermal.Summary, []thermal.DailyRow, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=30000000000")
	if err != nil {
		return thermal.Summary{}, nil, err
	}
	defer db.Close()
	_, _ = db.Exec("PRAGMA cache_size = -64000; PRAGMA mmap_size = 30000000000;")

	var hasModelUsage bool
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='model_usage'`).Scan(&hasModelUsage); err != nil {
		return thermal.Summary{}, nil, err
	}
	if !hasModelUsage {
		return thermal.Summary{}, nil, fmt.Errorf("zcode: model_usage table not found in %s", dbPath)
	}

	var summary thermal.Summary
	err = db.QueryRow(`
		SELECT
			COUNT(DISTINCT session_id),
			COALESCE(SUM(computed_total_tokens), 0),
			COALESCE(SUM(input_tokens), 0),
			COALESCE(SUM(output_tokens), 0),
			COALESCE(SUM(reasoning_tokens), 0),
			COALESCE(SUM(cache_creation_input_tokens + cache_read_input_tokens), 0)
		FROM model_usage
		WHERE status = 'completed'
	`).Scan(&summary.Sessions, &summary.LifetimeTokens, &summary.InputTokens,
		&summary.OutputTokens, &summary.ReasoningTokens, &summary.CacheTokens)
	if err != nil {
		return thermal.Summary{}, nil, err
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

	// Daily tokens from completed model requests. Turns come from completed
	// turn_usage rows (one per user-visible turn); when that table is absent,
	// fall back to completed request counts.
	var hasTurns bool
	_ = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='turn_usage'`).Scan(&hasTurns)

	turnExpr := `(SELECT COUNT(*) FROM model_usage m2
		WHERE m2.status = 'completed'
		AND date(m2.started_at / 1000, 'unixepoch', 'localtime') = date(model_usage.started_at / 1000, 'unixepoch', 'localtime'))`
	if hasTurns {
		turnExpr = `(SELECT COUNT(*) FROM turn_usage t
		WHERE t.status = 'completed'
		AND date(t.started_at / 1000, 'unixepoch', 'localtime') = date(model_usage.started_at / 1000, 'unixepoch', 'localtime'))`
	}
	rows, err := db.Query(`
		SELECT
			date(started_at / 1000, 'unixepoch', 'localtime') AS day,
			COALESCE(SUM(computed_total_tokens), 0),
			` + turnExpr + `
		FROM model_usage
		WHERE status = 'completed'
		GROUP BY day
		ORDER BY day
	`)
	if err != nil {
		return thermal.Summary{}, nil, err
	}
	defer rows.Close()

	var daily []thermal.DailyRow
	for rows.Next() {
		var r thermal.DailyRow
		if err := rows.Scan(&r.Day, &r.Tokens, &r.Turns); err != nil {
			return thermal.Summary{}, nil, err
		}
		daily = append(daily, r)
	}

	return summary, daily, nil
}
