package loaders

import (
	"database/sql"
	"sort"

	"github.com/jadmadi/thermal/internal/render"
	"github.com/jadmadi/thermal/internal/thermal"

	_ "modernc.org/sqlite"
)

// LoadOpenCodeData reads the OpenCode SQLite DB. OpenCode has pre-aggregated
// token columns on the session table (tokens_input, tokens_output, etc.) plus
// cost, agent, model, and code-change summaries (summary_additions/deletions/
// files). Using the session columns is faster than re-aggregating from
// message.data JSON and surfaces cost + code changes that were previously
// hidden.
func LoadOpenCodeData(dbPath string) (thermal.Summary, []thermal.DailyRow, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=30000000000")
	if err != nil {
		return thermal.Summary{}, nil, err
	}
	defer db.Close()
	_, _ = db.Exec("PRAGMA cache_size = -64000; PRAGMA mmap_size = 30000000000;")

	var summary thermal.Summary
	// Check whether the session table has the pre-agg token columns (newer
	// schemas). Fall back to message-level aggregation if not.
	var hasSessionCols bool
	err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('session') WHERE name = 'tokens_input'`).Scan(&hasSessionCols)
	if err != nil {
		return thermal.Summary{}, nil, err
	}

	if hasSessionCols {
		// Fast path: pre-aggregated session columns + cost + code changes.
		err = db.QueryRow(`
			SELECT
				COUNT(*),
				COALESCE(SUM(tokens_input + tokens_output + tokens_reasoning + tokens_cache_read + tokens_cache_write), 0),
				COALESCE(SUM(tokens_input), 0),
				COALESCE(SUM(tokens_output), 0),
				COALESCE(SUM(tokens_reasoning), 0),
				COALESCE(SUM(tokens_cache_read + tokens_cache_write), 0),
				COALESCE(SUM(cost), 0),
				COALESCE(SUM(summary_additions), 0),
				COALESCE(SUM(summary_deletions), 0),
				COALESCE(SUM(summary_files), 0),
				COALESCE(MAX(time_updated - time_created), 0)
			FROM session
		`).Scan(&summary.Sessions, &summary.LifetimeTokens, &summary.InputTokens,
			&summary.OutputTokens, &summary.ReasoningTokens, &summary.CacheTokens,
			&summary.Cost, &summary.LinesAdded, &summary.LinesDeleted,
			&summary.FilesTouched, &summary.LongestSessionMs)
		if err != nil {
			return thermal.Summary{}, nil, err
		}

		// Agent breakdown from session.agent column.
		agentRows, err := db.Query(`SELECT agent, COUNT(*) FROM session WHERE agent != '' GROUP BY agent`)
		if err == nil {
			summary.AgentBreakdown = make(map[string]int)
			for agentRows.Next() {
				var agent string
				var n int
				agentRows.Scan(&agent, &n)
				summary.AgentBreakdown[agent] = n
			}
			agentRows.Close()
		}

		// Daily aggregation from session pre-agg columns (one row per session,
		// far fewer rows than message-level).
		rows, err := db.Query(`
			SELECT
				date(time_created / 1000, 'unixepoch', 'localtime') AS day,
				COALESCE(SUM(tokens_input + tokens_output + tokens_reasoning + tokens_cache_read + tokens_cache_write), 0),
				COUNT(*)
			FROM session
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

	// Fallback: older schema without session token columns — aggregate from
	// message.data JSON (same as MiMo).
	return loadMessageLevelData(db)
}

// LoadMiMoCodeData reads the MiMoCode SQLite DB. MiMo has no pre-aggregated
// session token columns, so we aggregate from message.data.tokens JSON. We
// also surface session.summary_additions/deletions/files (code changes) and
// message.data.agent (agent mode distribution) that were previously hidden.
func LoadMiMoCodeData(dbPath string) (thermal.Summary, []thermal.DailyRow, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=30000000000")
	if err != nil {
		return thermal.Summary{}, nil, err
	}
	defer db.Close()
	_, _ = db.Exec("PRAGMA cache_size = -64000; PRAGMA mmap_size = 30000000000;")

	summary, daily, err := loadMessageLevelData(db)
	if err != nil {
		return summary, daily, err
	}

	// Code-change analytics from session summary columns.
	_ = db.QueryRow(`
		SELECT COALESCE(SUM(summary_additions), 0), COALESCE(SUM(summary_deletions), 0), COALESCE(SUM(summary_files), 0)
		FROM session WHERE summary_additions IS NOT NULL
	`).Scan(&summary.LinesAdded, &summary.LinesDeleted, &summary.FilesTouched)

	// Agent mode distribution from message.data.agent.
	agentRows, err := db.Query(`SELECT json_extract(data, '$.agent'), COUNT(*) FROM message WHERE data LIKE '%"assistant"%' AND json_extract(data, '$.role') = 'assistant' AND json_extract(data, '$.agent') != '' GROUP BY 1`)
	if err == nil {
		summary.AgentBreakdown = make(map[string]int)
		for agentRows.Next() {
			var agent string
			var n int
			agentRows.Scan(&agent, &n)
			summary.AgentBreakdown[agent] = n
		}
		agentRows.Close()
	}

	return summary, daily, nil
}

// loadMessageLevelData aggregates token metrics from the message table's
// data JSON column. Shared by MiMo (always) and OpenCode (fallback for older
// schemas without session-level token columns).
func loadMessageLevelData(db *sql.DB) (thermal.Summary, []thermal.DailyRow, error) {
	var summary thermal.Summary
	err := db.QueryRow(`
		SELECT
			COUNT(DISTINCT session_id),
			COALESCE(SUM(
				COALESCE(CAST(json_extract(data, '$.tokens.input') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.output') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.reasoning') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.cache.read') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.cache.write') AS INTEGER), 0)
			), 0),
			COALESCE(SUM(COALESCE(CAST(json_extract(data, '$.tokens.input') AS INTEGER), 0)), 0),
			COALESCE(SUM(COALESCE(CAST(json_extract(data, '$.tokens.output') AS INTEGER), 0)), 0),
			COALESCE(SUM(COALESCE(CAST(json_extract(data, '$.tokens.reasoning') AS INTEGER), 0)), 0),
			COALESCE(SUM(
				COALESCE(CAST(json_extract(data, '$.tokens.cache.read') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.cache.write') AS INTEGER), 0)
			), 0),
			COALESCE(SUM(COALESCE(CAST(json_extract(data, '$.cost') AS REAL), 0)), 0)
		FROM message
		WHERE data LIKE '%"assistant"%' AND json_extract(data, '$.role') = 'assistant'
	`).Scan(&summary.Sessions, &summary.LifetimeTokens, &summary.InputTokens,
		&summary.OutputTokens, &summary.ReasoningTokens, &summary.CacheTokens, &summary.Cost)
	if err != nil {
		return thermal.Summary{}, nil, err
	}

	db.QueryRow(`SELECT COALESCE(MAX(time_updated - time_created), 0) FROM session`).
		Scan(&summary.LongestSessionMs)

	rows, err := db.Query(`
		SELECT
			date(time_created / 1000, 'unixepoch', 'localtime') AS day,
			SUM(
				COALESCE(CAST(json_extract(data, '$.tokens.input') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.output') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.reasoning') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.cache.read') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(data, '$.tokens.cache.write') AS INTEGER), 0)
			) AS tokens,
			COUNT(*) AS turns
		FROM message
		WHERE data LIKE '%"assistant"%' AND json_extract(data, '$.role') = 'assistant'
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

func LoadDevinData(dbPath string) (thermal.Summary, []thermal.DailyRow, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=30000000000")
	if err != nil {
		return thermal.Summary{}, nil, err
	}
	defer db.Close()
	_, _ = db.Exec("PRAGMA cache_size = -64000; PRAGMA mmap_size = 30000000000;")

	// Cheap invalidation probes (<2ms via PK/stat indexes). message_nodes is
	// append-only in practice, so MAX(row_id) covers new data; the visible
	// session count covers hidden/unhidden toggles. If both match the on-disk
	// cache, skip the ~11s full scan and return the cached snapshot.
	var maxRowID int64
	var sessionCount int
	if err := db.QueryRow(`SELECT MAX(row_id) FROM message_nodes`).Scan(&maxRowID); err != nil {
		return thermal.Summary{}, nil, err
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE hidden = 0`).Scan(&sessionCount); err != nil {
		return thermal.Summary{}, nil, err
	}
	if c, ok := loadDevinCache(); ok && c.MaxRowID == maxRowID && c.SessionCount == sessionCount {
		return c.Summary, c.Daily, nil
	}

	// Longest session duration (created_at/last_activity_at are in seconds;
	// LongestSessionMs is expected in milliseconds, so *1000).
	var longestSec int64
	if err := db.QueryRow(`SELECT COALESCE(MAX(last_activity_at - created_at), 0) FROM sessions WHERE hidden = 0`).Scan(&longestSec); err != nil {
		return thermal.Summary{}, nil, err
	}

	// If session count is unchanged and new messages were simply appended (maxRowID > c.MaxRowID),
	// perform a lightning-fast delta scan on only the new message_nodes rows via PK index.
	if c, ok := loadDevinCache(); ok && c.SessionCount == sessionCount && maxRowID > c.MaxRowID && c.MaxRowID > 0 {
		c.Summary.Sessions = sessionCount
		c.Summary.LongestSessionMs = longestSec * 1000

		deltaRows, err := db.Query(`
			SELECT created_at,
			       json_extract(chat_message, '$.metadata.metrics.input_tokens'),
			       json_extract(chat_message, '$.metadata.metrics.output_tokens'),
			       json_extract(chat_message, '$.metadata.metrics.cache_read_tokens'),
			       json_extract(chat_message, '$.metadata.metrics.cache_creation_tokens')
			FROM message_nodes
			WHERE row_id > ? AND chat_message LIKE '%"assistant"%' AND json_extract(chat_message, '$.role') = 'assistant'
		`, c.MaxRowID)
		if err == nil {
			defer deltaRows.Close()
			type dayAgg struct {
				tokens int64
				turns  int
			}
			byDay := make(map[string]*dayAgg)
			for _, r := range c.Daily {
				byDay[r.Day] = &dayAgg{
					tokens: r.Tokens,
					turns:  r.Turns,
				}
			}

			for deltaRows.Next() {
				var createdAt int64
				var inTok, outTok, cacheRead, cacheCreate sql.NullInt64
				if err := deltaRows.Scan(&createdAt, &inTok, &outTok, &cacheRead, &cacheCreate); err != nil {
					break
				}
				day := thermal.UnixDay(createdAt)
				agg := byDay[day]
				if agg == nil {
					agg = &dayAgg{}
					byDay[day] = agg
				}
				deltaTok := inTok.Int64 + outTok.Int64 + cacheRead.Int64 + cacheCreate.Int64
				agg.tokens += deltaTok
				agg.turns++

				c.Summary.InputTokens += inTok.Int64
				c.Summary.OutputTokens += outTok.Int64
				c.Summary.CacheTokens += cacheRead.Int64 + cacheCreate.Int64
				c.Summary.LifetimeTokens += deltaTok
			}

			if err := deltaRows.Err(); err == nil {
				var daily []thermal.DailyRow
				for day, agg := range byDay {
					daily = append(daily, thermal.DailyRow{
						Day:    day,
						Tokens: agg.tokens,
						Turns:  agg.turns,
					})
				}
				sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })
				c.Daily = daily
				c.MaxRowID = maxRowID
				saveDevinCache(c)
				return c.Summary, c.Daily, nil
			}
		}
	}

	var summary thermal.Summary
	summary.Sessions = sessionCount
	summary.LongestSessionMs = longestSec * 1000

	// Single pass over message_nodes: stream rows into Go and aggregate
	// there, so we can drive a live progress bar from the row counter. The
	// total row count (instant via internal stats) is the denominator. Using
	// per-message created_at means a long agentic session spanning midnight
	// contributes to each day it generated tokens, not just the day it
	// started.
	var totalCount int64
	if err := db.QueryRow(`SELECT COUNT(*) FROM message_nodes`).Scan(&totalCount); err != nil {
		return thermal.Summary{}, nil, err
	}

	rows, err := db.Query(`
		SELECT created_at,
		       json_extract(chat_message, '$.metadata.metrics.input_tokens'),
		       json_extract(chat_message, '$.metadata.metrics.output_tokens'),
		       json_extract(chat_message, '$.metadata.metrics.cache_read_tokens'),
		       json_extract(chat_message, '$.metadata.metrics.cache_creation_tokens')
		FROM message_nodes
		WHERE chat_message LIKE '%"assistant"%' AND json_extract(chat_message, '$.role') = 'assistant'
	`)
	if err != nil {
		return thermal.Summary{}, nil, err
	}
	defer rows.Close()

	progress := render.NewProgress("Devin", totalCount)
	progress.Start()

	type dayAgg struct {
		inTok, outTok, cacheTok int64
		turns                   int
	}
	byDay := make(map[string]*dayAgg)
	var scanned int64
	for rows.Next() {
		var createdAt int64
		var inTok, outTok, cacheRead, cacheCreate sql.NullInt64
		if err := rows.Scan(&createdAt, &inTok, &outTok, &cacheRead, &cacheCreate); err != nil {
			progress.Done()
			return thermal.Summary{}, nil, err
		}
		day := thermal.UnixDay(createdAt)
		agg := byDay[day]
		if agg == nil {
			agg = &dayAgg{}
			byDay[day] = agg
		}
		agg.inTok += inTok.Int64
		agg.outTok += outTok.Int64
		agg.cacheTok += cacheRead.Int64 + cacheCreate.Int64
		agg.turns++
		scanned++
		if scanned%2000 == 0 {
			progress.Increment(2000)
		}
	}
	progress.Increment(scanned % 2000)
	progress.Done()

	if err := rows.Err(); err != nil {
		return thermal.Summary{}, nil, err
	}

	var daily []thermal.DailyRow
	for day, agg := range byDay {
		summary.InputTokens += agg.inTok
		summary.OutputTokens += agg.outTok
		summary.CacheTokens += agg.cacheTok
		daily = append(daily, thermal.DailyRow{
			Day:    day,
			Tokens: agg.inTok + agg.outTok + agg.cacheTok,
			Turns:  agg.turns,
		})
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })
	// Devin folds reasoning into output_tokens; no separate reasoning field.
	summary.ReasoningTokens = 0
	summary.LifetimeTokens = summary.InputTokens + summary.OutputTokens + summary.CacheTokens

	// Fallback: if message_nodes has no assistant metrics (e.g. very old DB
	// schema), fall back to per-session-day counts so the heatmap still works.
	if len(daily) == 0 {
		fbRows, fbErr := db.Query(`
			SELECT date(created_at, 'unixepoch', 'localtime') AS day, COUNT(*) AS turns
			FROM sessions
			WHERE hidden = 0
			GROUP BY day
			ORDER BY day
		`)
		if fbErr != nil {
			return summary, nil, nil
		}
		defer fbRows.Close()
		for fbRows.Next() {
			var r thermal.DailyRow
			if err := fbRows.Scan(&r.Day, &r.Turns); err != nil {
				return summary, nil, nil
			}
			r.Tokens = int64(r.Turns)
			daily = append(daily, r)
		}
		summary.LifetimeTokens = int64(summary.Sessions)
	}

	saveDevinCache(DevinCache{
		MaxRowID:     maxRowID,
		SessionCount: sessionCount,
		Summary:      summary,
		Daily:        daily,
	})

	return summary, daily, nil
}
