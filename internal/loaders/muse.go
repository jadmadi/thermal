package loaders

import (
	"database/sql"

	"github.com/jadmadi/thermal/internal/thermal"

	_ "modernc.org/sqlite"
)

// LoadMuseData reads the Meta Muse session index
// (~/.local/share/muse/session-index.db). The index records session identity,
// prompt counts, model ids, and microsecond timestamps, but no token or cost
// telemetry, so this is an activity-only loader: prompt counts stand in for
// activity the way message counts do for command-code. When local sessions
// accumulate model-call frames, a token upgrade can read per-session
// session.jsonl logs.
func LoadMuseData(dbPath string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=30000000000")
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	defer db.Close()
	_, _ = db.Exec("PRAGMA cache_size = -64000; PRAGMA mmap_size = 30000000000;")

	var summary thermal.Summary
	err = db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(prompt_count), 0),
			COALESCE(MAX(updated_at_us - created_at_us) / 1000, 0)
		FROM sessions
	`).Scan(&summary.Sessions, &summary.LifetimeTokens, &summary.LongestSessionMs)
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}

	// Model distribution from indexed model ids.
	if modelRows, err := db.Query(`SELECT model_id, COUNT(*) FROM sessions WHERE model_id IS NOT NULL AND model_id != '' GROUP BY model_id`); err == nil {
		summary.ModelBreakdown = make(map[string]int64)
		for modelRows.Next() {
			var model string
			var n int64
			modelRows.Scan(&model, &n)
			summary.ModelBreakdown[model] = n
		}
		modelRows.Close()
	}

	// Daily prompt activity by last-update day. Days without prompts are
	// omitted so they never count as active.
	rows, err := db.Query(`
		SELECT
			date(updated_at_us / 1000000, 'unixepoch', 'localtime') AS day,
			COALESCE(SUM(prompt_count), 0)
		FROM sessions
		GROUP BY day
		HAVING SUM(prompt_count) > 0
		ORDER BY day
	`)
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	defer rows.Close()

	var daily []thermal.DailyRow
	for rows.Next() {
		var r thermal.DailyRow
		if err := rows.Scan(&r.Day, &r.Tokens); err != nil {
			return thermal.Summary{}, nil, nil, err
		}
		r.Turns = int(r.Tokens)
		daily = append(daily, r)
	}

	return summary, daily, nil, nil
}
