package loaders

import (
	"database/sql"
	"sort"

	"github.com/jadmadi/thermal/internal/render"
	"github.com/jadmadi/thermal/internal/thermal"

	_ "modernc.org/sqlite"
)

// LoadOpenCodeData reads the OpenCode SQLite DB. Since the v2 storage
// migration, new sessions land in session_v2 and the legacy session table
// stops receiving writes, so session_v2 is the primary source. Both tables
// carry pre-aggregated token columns (tokens_input, tokens_output, etc.) plus
// cost, agent, model, and code-change summaries (summary_additions/deletions/
// files). Legacy rows whose ids are absent from session_v2 are folded in so
// pre-migration history is not dropped. For DBs predating session_v2, fall
// back to the session table, and to message.data JSON aggregation when the
// session table has no token columns.
func LoadOpenCodeData(dbPath string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=30000000000")
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	defer db.Close()
	_, _ = db.Exec("PRAGMA cache_size = -64000; PRAGMA mmap_size = 30000000000;")

	var summary thermal.Summary

	// OpenCode v2 moved sessions into session_v2; the legacy session table
	// keeps only pre-migration rows. Read v2 when present.
	var hasV2 bool
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='session_v2'`).Scan(&hasV2); err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	if hasV2 {
		src := "(" + opencodeV2Source(db) + ")"

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
			FROM `+src).Scan(&summary.Sessions, &summary.LifetimeTokens, &summary.InputTokens,
			&summary.OutputTokens, &summary.ReasoningTokens, &summary.CacheTokens,
			&summary.Cost, &summary.LinesAdded, &summary.LinesDeleted,
			&summary.FilesTouched, &summary.LongestSessionMs)
		if err != nil {
			return thermal.Summary{}, nil, nil, err
		}

		// Agent breakdown from session.agent column.
		agentRows, err := db.Query(`SELECT agent, COUNT(*) FROM ` + src + ` WHERE agent != '' GROUP BY agent`)
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

		// Daily aggregation from pre-agg columns, one row per session,
		// project, and model, far fewer rows than message-level.
		rows, err := db.Query(`
			SELECT
				date(time_created / 1000, 'unixepoch', 'localtime') AS day,
				project,
				COALESCE(model, '') AS model,
				COALESCE(SUM(tokens_input), 0),
				COALESCE(SUM(tokens_output), 0),
				COALESCE(SUM(tokens_reasoning), 0),
				COALESCE(SUM(tokens_cache_read), 0),
				COALESCE(SUM(tokens_cache_write), 0),
				COALESCE(SUM(cost), 0),
				COUNT(*),
				COALESCE(SUM(tokens_input + tokens_output + tokens_reasoning + tokens_cache_read + tokens_cache_write), 0)
			FROM ` + src + `
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
		return summary, daily, projects, nil
	}

	// Check whether the session table has the pre-agg token columns (newer
	// schemas). Fall back to message-level aggregation if not.
	// Check whether the session table has the pre-agg token columns (newer
	// schemas). Fall back to message-level aggregation if not.
	var hasSessionCols bool
	err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('session') WHERE name = 'tokens_input'`).Scan(&hasSessionCols)
	if err != nil {
		return thermal.Summary{}, nil, nil, err
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
			return thermal.Summary{}, nil, nil, err
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

		// Daily aggregation from session pre-agg columns, one row per session,
		// project, and model, far fewer rows than message-level. Older schemas
		// may lack the model or directory columns; empty literals keep the
		// query valid.
		legacyModel := "''"
		if hasColumn(db, "session", "model") {
			legacyModel = modelIDExpr("model")
		}
		rows, err := db.Query(`
			SELECT
				date(time_created / 1000, 'unixepoch', 'localtime') AS day,
				` + sessionProjectExpr(db, "session") + ` AS project,
				` + legacyModel + ` AS model,
				COALESCE(SUM(tokens_input), 0),
				COALESCE(SUM(tokens_output), 0),
				COALESCE(SUM(tokens_reasoning), 0),
				COALESCE(SUM(tokens_cache_read), 0),
				COALESCE(SUM(tokens_cache_write), 0),
				COALESCE(SUM(cost), 0),
				COUNT(*),
				COALESCE(SUM(tokens_input + tokens_output + tokens_reasoning + tokens_cache_read + tokens_cache_write), 0)
			FROM session
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
		return summary, daily, projects, nil
	}

	// Fallback: older schema without session token columns — aggregate from
	// message.data JSON (same as MiMo).
	return loadMessageLevelData(db)
}

// modelIDExpr builds a SQL expression that reads a model id from a column
// holding either a JSON object ({"id": ...}) or a plain string. json_extract
// raises an error on non-JSON text, so json_valid gates it, and a NULL column
// becomes an empty string.
func modelIDExpr(col string) string {
	return `COALESCE(json_extract(CASE WHEN json_valid(` + col + `) THEN ` + col + ` END, '$.id'), ` + col + `, '')`
}

// opencodeV2Source builds a session row source covering session_v2 plus any
// legacy session rows that were never migrated (ids absent from session_v2).
// Child/subagent sessions carry their own token totals, so every row counts
// real usage exactly once. The legacy arm is skipped when the session table
// lacks an id or token columns, and the model and directory columns are
// replaced with empty literals when absent so the UNION stays valid.
func opencodeV2Source(db *sql.DB) string {
	const cols = `SELECT id, tokens_input, tokens_output, tokens_reasoning,
			tokens_cache_read, tokens_cache_write, cost,
			summary_additions, summary_deletions, summary_files,
			agent, time_created, time_updated,`
	v2Model := modelIDExpr("model")
	if !hasColumn(db, "session_v2", "model") {
		v2Model = "''"
	}
	v2Project := sessionProjectExpr(db, "session_v2")
	source := cols + ` ` + v2Model + ` AS model, ` + v2Project + ` AS project
		FROM session_v2`
	var colsSeen int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('session') WHERE name IN ('id', 'tokens_input')`).Scan(&colsSeen); err == nil && colsSeen == 2 {
		legacyModel := modelIDExpr("model")
		if !hasColumn(db, "session", "model") {
			legacyModel = "''"
		}
		legacyProject := sessionProjectExpr(db, "session")
		source += `
		UNION ALL ` + cols + ` ` + legacyModel + ` AS model, ` + legacyProject + ` AS project
		FROM session
		WHERE id NOT IN (SELECT id FROM session_v2)`
	}
	return source
}

// hasColumn reports whether a table carries the named column. A missing table
// also reports false.
func hasColumn(db *sql.DB, table, column string) bool {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`, table, column).Scan(&n); err != nil {
		return false
	}
	return n > 0
}

// sessionProjectExpr reads a session directory as a project path, falling back
// to an empty literal when the column is absent or blank.
func sessionProjectExpr(db *sql.DB, table string) string {
	if !hasColumn(db, table, "directory") {
		return "''"
	}
	return `COALESCE(NULLIF(directory, ''), '')`
}

// projectDayKey identifies one day of one project.
type projectDayKey struct{ day, project string }

// foldDayModelProjectRows turns a day, project, and model aggregation result
// into DailyRows plus ProjectDays. Expected column order: day, project, model,
// input, output, reasoning, cache read, cache write, cost, turns, total. The
// explicit total is authoritative because some sources pack tokens differently
// (ZCode counts cache reads inside input, for example). Cache reads and writes
// stay separate for pricing, and their sum feeds DailyRow.Cache. Project paths
// are normalized to their git root, and blank projects are skipped.
func foldDayModelProjectRows(rows *sql.Rows) ([]thermal.DailyRow, []thermal.ProjectDay, error) {
	byDay := make(map[string]*thermal.DailyRow)
	byProjectDay := make(map[projectDayKey]*thermal.ProjectDay)
	modelsByDay := make(map[string]map[string]thermal.ModelTokens)
	modelsByProjectDay := make(map[projectDayKey]map[string]thermal.ModelTokens)
	var dayOrder []string
	var projectOrder []projectDayKey

	for rows.Next() {
		var day, project, model string
		var input, output, reasoning, cacheRead, cacheWrite int64
		var cost float64
		var turns int
		var total int64
		if err := rows.Scan(&day, &project, &model, &input, &output, &reasoning, &cacheRead, &cacheWrite, &cost, &turns, &total); err != nil {
			return nil, nil, err
		}
		model = modelName(model)
		projectKey := thermal.ProjectKey(project)

		row := byDay[day]
		if row == nil {
			row = &thermal.DailyRow{Day: day}
			byDay[day] = row
			dayOrder = append(dayOrder, day)
		}
		row.Input += input
		row.Output += output
		row.Reasoning += reasoning
		row.Cache += cacheRead + cacheWrite
		row.Tokens += total
		row.Cost += cost
		row.Turns += turns

		if projectKey != "" {
			key := projectDayKey{day, projectKey}
			pd := byProjectDay[key]
			if pd == nil {
				pd = &thermal.ProjectDay{Project: projectKey, Day: day}
				byProjectDay[key] = pd
				projectOrder = append(projectOrder, key)
			}
			pd.Input += input
			pd.Output += output
			pd.Reasoning += reasoning
			pd.CacheRead += cacheRead
			pd.CacheWrite += cacheWrite
			pd.Tokens += total
			pd.Cost += cost
			pd.Turns += turns
		}

		if model != "" {
			if modelsByDay[day] == nil {
				modelsByDay[day] = make(map[string]thermal.ModelTokens)
			}
			modelsByDay[day][model] = modelsByDay[day][model].Add(thermal.ModelTokens{
				Input:      input,
				Output:     output,
				Reasoning:  reasoning,
				CacheRead:  cacheRead,
				CacheWrite: cacheWrite,
			})
			if projectKey != "" {
				key := projectDayKey{day, projectKey}
				if modelsByProjectDay[key] == nil {
					modelsByProjectDay[key] = make(map[string]thermal.ModelTokens)
				}
				modelsByProjectDay[key][model] = modelsByProjectDay[key][model].Add(thermal.ModelTokens{
					Input:      input,
					Output:     output,
					Reasoning:  reasoning,
					CacheRead:  cacheRead,
					CacheWrite: cacheWrite,
				})
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	daily := make([]thermal.DailyRow, 0, len(dayOrder))
	for _, day := range dayOrder {
		row := byDay[day]
		row.Models = modelsByDay[day]
		daily = append(daily, *row)
	}

	projects := make([]thermal.ProjectDay, 0, len(projectOrder))
	for _, key := range projectOrder {
		pd := byProjectDay[key]
		pd.Models = modelsByProjectDay[key]
		projects = append(projects, *pd)
	}
	return daily, projects, nil
}

// LoadMiMoCodeData reads the MiMoCode SQLite DB. MiMo has no pre-aggregated
// session token columns, so we aggregate from message.data.tokens JSON. We
// also surface session.summary_additions/deletions/files (code changes) and
// message.data.agent (agent mode distribution) that were previously hidden.
func LoadMiMoCodeData(dbPath string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=30000000000")
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	defer db.Close()
	_, _ = db.Exec("PRAGMA cache_size = -64000; PRAGMA mmap_size = 30000000000;")

	summary, daily, projects, err := loadMessageLevelData(db)
	if err != nil {
		return summary, daily, projects, err
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

	return summary, daily, projects, nil
}

// loadMessageLevelData aggregates token metrics from the message table's
// data JSON column. Shared by MiMo (always) and OpenCode (fallback for older
// schemas without session-level token columns).
func loadMessageLevelData(db *sql.DB) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
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
		return thermal.Summary{}, nil, nil, err
	}

	db.QueryRow(`SELECT COALESCE(MAX(time_updated - time_created), 0) FROM session`).
		Scan(&summary.LongestSessionMs)

	// Project attribution comes from the session row when the schema carries a
	// directory for it. Columns are qualified because session also has
	// time_created.
	projectExpr := "''"
	from := "FROM message"
	if hasColumn(db, "session", "directory") {
		projectExpr = "COALESCE(NULLIF(s.directory, ''), '')"
		from = "FROM message LEFT JOIN session s ON s.id = message.session_id"
	}
	rows, err := db.Query(`
		SELECT
			date(message.time_created / 1000, 'unixepoch', 'localtime') AS day,
			` + projectExpr + ` AS project,
			COALESCE(
				NULLIF(NULLIF(json_extract(message.data, '$.modelID'), ''), '<synthetic>'),
				NULLIF(json_extract(message.data, '$.model'), ''),
				''
			) AS model,
			COALESCE(SUM(CAST(json_extract(message.data, '$.tokens.input') AS INTEGER)), 0),
			COALESCE(SUM(CAST(json_extract(message.data, '$.tokens.output') AS INTEGER)), 0),
			COALESCE(SUM(CAST(json_extract(message.data, '$.tokens.reasoning') AS INTEGER)), 0),
			COALESCE(SUM(CAST(json_extract(message.data, '$.tokens.cache.read') AS INTEGER)), 0),
			COALESCE(SUM(CAST(json_extract(message.data, '$.tokens.cache.write') AS INTEGER)), 0),
			COALESCE(SUM(CAST(json_extract(message.data, '$.cost') AS REAL)), 0),
			COUNT(*),
			COALESCE(SUM(
				COALESCE(CAST(json_extract(message.data, '$.tokens.input') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(message.data, '$.tokens.output') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(message.data, '$.tokens.reasoning') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(message.data, '$.tokens.cache.read') AS INTEGER), 0) +
				COALESCE(CAST(json_extract(message.data, '$.tokens.cache.write') AS INTEGER), 0)
			), 0)
		` + from + `
		WHERE message.data LIKE '%"assistant"%' AND json_extract(message.data, '$.role') = 'assistant'
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

	return summary, daily, projects, nil
}

func LoadDevinData(dbPath string) (thermal.Summary, []thermal.DailyRow, []thermal.ProjectDay, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro&_pragma=cache_size=-64000&_pragma=mmap_size=30000000000")
	if err != nil {
		return thermal.Summary{}, nil, nil, err
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
		return thermal.Summary{}, nil, nil, err
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE hidden = 0`).Scan(&sessionCount); err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	if c, ok := loadDevinCache(); ok && c.MaxRowID == maxRowID && c.SessionCount == sessionCount {
		return c.Summary, c.Daily, c.Projects, nil
	}

	// Longest session duration (created_at/last_activity_at are in seconds;
	// LongestSessionMs is expected in milliseconds, so *1000).
	var longestSec int64
	if err := db.QueryRow(`SELECT COALESCE(MAX(last_activity_at - created_at), 0) FROM sessions WHERE hidden = 0`).Scan(&longestSec); err != nil {
		return thermal.Summary{}, nil, nil, err
	}

	// Project attribution needs the session working directory, which older
	// Devin schemas may not carry.
	projectSelect, projectJoin := "''", ""
	if hasColumn(db, "sessions", "working_directory") {
		projectSelect = "COALESCE(s.working_directory, '')"
		projectJoin = " LEFT JOIN sessions s ON s.id = m.session_id"
	}

	// If session count is unchanged and new messages were simply appended (maxRowID > c.MaxRowID),
	// perform a lightning-fast delta scan on only the new message_nodes rows via PK index.
	if c, ok := loadDevinCache(); ok && c.SessionCount == sessionCount && maxRowID > c.MaxRowID && c.MaxRowID > 0 {
		c.Summary.Sessions = sessionCount
		c.Summary.LongestSessionMs = longestSec * 1000

		deltaRows, err := db.Query(`
			SELECT m.created_at,
			       `+projectSelect+`,
			       json_extract(m.chat_message, '$.metadata.metrics.input_tokens'),
			       json_extract(m.chat_message, '$.metadata.metrics.output_tokens'),
			       json_extract(m.chat_message, '$.metadata.metrics.cache_read_tokens'),
			       json_extract(m.chat_message, '$.metadata.metrics.cache_creation_tokens')
			FROM message_nodes m`+projectJoin+`
			WHERE m.row_id > ? AND m.chat_message LIKE '%"assistant"%' AND json_extract(m.chat_message, '$.role') = 'assistant'
		`, c.MaxRowID)
		if err == nil {
			defer deltaRows.Close()
			type dayAgg struct {
				tokens   int64
				inTok    int64
				outTok   int64
				cacheTok int64
				turns    int
			}
			byDay := make(map[string]*dayAgg)
			modelsByDay := make(map[string]map[string]thermal.ModelTokens)
			for _, r := range c.Daily {
				byDay[r.Day] = &dayAgg{
					tokens:   r.Tokens,
					inTok:    r.Input,
					outTok:   r.Output,
					cacheTok: r.Cache,
					turns:    r.Turns,
				}
				if len(r.Models) > 0 {
					modelsByDay[r.Day] = r.Models
				}
			}
			byProjectDay := make(map[projectDayKey]*thermal.ProjectDay)
			for _, p := range c.Projects {
				byProjectDay[projectDayKey{p.Day, p.Project}] = &p
			}

			for deltaRows.Next() {
				var createdAt int64
				var workingDir string
				var inTok, outTok, cacheRead, cacheCreate sql.NullInt64
				if err := deltaRows.Scan(&createdAt, &workingDir, &inTok, &outTok, &cacheRead, &cacheCreate); err != nil {
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
				agg.inTok += inTok.Int64
				agg.outTok += outTok.Int64
				agg.cacheTok += cacheRead.Int64 + cacheCreate.Int64
				agg.turns++

				if project := thermal.ProjectKey(workingDir); project != "" {
					key := projectDayKey{day, project}
					pd := byProjectDay[key]
					if pd == nil {
						pd = &thermal.ProjectDay{Project: project, Day: day}
						byProjectDay[key] = pd
					}
					pd.Tokens += deltaTok
					pd.Input += inTok.Int64
					pd.Output += outTok.Int64
					pd.CacheRead += cacheRead.Int64
					pd.CacheWrite += cacheCreate.Int64
					pd.Turns++
				}

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
						Input:  agg.inTok,
						Output: agg.outTok,
						Cache:  agg.cacheTok,
						Turns:  agg.turns,
						Models: modelsByDay[day],
					})
				}
				sort.Slice(daily, func(i, j int) bool { return daily[i].Day < daily[j].Day })
				c.Daily = daily
				c.Projects = sortedProjects(byProjectDay)
				c.MaxRowID = maxRowID
				saveDevinCache(c)
				return c.Summary, c.Daily, c.Projects, nil
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
		return thermal.Summary{}, nil, nil, err
	}

	rows, err := db.Query(`
		SELECT m.created_at,
		       ` + projectSelect + `,
		       json_extract(m.chat_message, '$.metadata.metrics.input_tokens'),
		       json_extract(m.chat_message, '$.metadata.metrics.output_tokens'),
		       json_extract(m.chat_message, '$.metadata.metrics.cache_read_tokens'),
		       json_extract(m.chat_message, '$.metadata.metrics.cache_creation_tokens')
		FROM message_nodes m` + projectJoin + `
		WHERE m.chat_message LIKE '%"assistant"%' AND json_extract(m.chat_message, '$.role') = 'assistant'
	`)
	if err != nil {
		return thermal.Summary{}, nil, nil, err
	}
	defer rows.Close()

	progress := render.NewProgress("Devin", totalCount)
	progress.Start()

	type dayAgg struct {
		inTok, outTok, cacheTok int64
		turns                   int
	}
	byDay := make(map[string]*dayAgg)
	byProjectDay := make(map[projectDayKey]*thermal.ProjectDay)
	var scanned int64
	for rows.Next() {
		var createdAt int64
		var workingDir string
		var inTok, outTok, cacheRead, cacheCreate sql.NullInt64
		if err := rows.Scan(&createdAt, &workingDir, &inTok, &outTok, &cacheRead, &cacheCreate); err != nil {
			progress.Done()
			return thermal.Summary{}, nil, nil, err
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

		if project := thermal.ProjectKey(workingDir); project != "" {
			key := projectDayKey{day, project}
			pd := byProjectDay[key]
			if pd == nil {
				pd = &thermal.ProjectDay{Project: project, Day: day}
				byProjectDay[key] = pd
			}
			pd.Input += inTok.Int64
			pd.Output += outTok.Int64
			pd.CacheRead += cacheRead.Int64
			pd.CacheWrite += cacheCreate.Int64
			pd.Tokens += inTok.Int64 + outTok.Int64 + cacheRead.Int64 + cacheCreate.Int64
			pd.Turns++
		}

		scanned++
		if scanned%2000 == 0 {
			progress.Increment(2000)
		}
	}
	progress.Increment(scanned % 2000)
	progress.Done()

	if err := rows.Err(); err != nil {
		return thermal.Summary{}, nil, nil, err
	}

	var daily []thermal.DailyRow
	for day, agg := range byDay {
		summary.InputTokens += agg.inTok
		summary.OutputTokens += agg.outTok
		summary.CacheTokens += agg.cacheTok
		daily = append(daily, thermal.DailyRow{
			Day:    day,
			Tokens: agg.inTok + agg.outTok + agg.cacheTok,
			Input:  agg.inTok,
			Output: agg.outTok,
			Cache:  agg.cacheTok,
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
			return summary, nil, nil, nil
		}
		defer fbRows.Close()
		for fbRows.Next() {
			var r thermal.DailyRow
			if err := fbRows.Scan(&r.Day, &r.Turns); err != nil {
				return summary, nil, nil, nil
			}
			r.Tokens = int64(r.Turns)
			daily = append(daily, r)
		}
		summary.LifetimeTokens = int64(summary.Sessions)
	}

	projects := sortedProjects(byProjectDay)
	saveDevinCache(DevinCache{
		MaxRowID:     maxRowID,
		SessionCount: sessionCount,
		Summary:      summary,
		Daily:        daily,
		Projects:     projects,
	})

	return summary, daily, projects, nil
}

// sortedProjects flattens a project-day map into a stable slice.
func sortedProjects(byProjectDay map[projectDayKey]*thermal.ProjectDay) []thermal.ProjectDay {
	if len(byProjectDay) == 0 {
		return nil
	}
	projects := make([]thermal.ProjectDay, 0, len(byProjectDay))
	for _, pd := range byProjectDay {
		projects = append(projects, *pd)
	}
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Day != projects[j].Day {
			return projects[i].Day < projects[j].Day
		}
		return projects[i].Project < projects[j].Project
	})
	return projects
}
