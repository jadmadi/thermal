package loaders

import (
	"database/sql"

	"thermal/internal/thermal"

	_ "modernc.org/sqlite"
)

func LoadSqliteData(dbPath string) (thermal.Summary, []thermal.DailyRow, error) {
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return thermal.Summary{}, nil, err
	}
	defer db.Close()

	var summary thermal.Summary
	err = db.QueryRow(`
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
		WHERE json_extract(data, '$.role') = 'assistant'
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
		WHERE json_extract(data, '$.role') = 'assistant'
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
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return thermal.Summary{}, nil, err
	}
	defer db.Close()

	var summary thermal.Summary
	err = db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&summary.Sessions)
	if err != nil {
		return thermal.Summary{}, nil, err
	}
	summary.LifetimeTokens = int64(summary.Sessions)

	rows, err := db.Query(`
		SELECT
			date(created_at, 'unixepoch', 'localtime') AS day,
			COUNT(*) AS turns
		FROM sessions
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
		if err := rows.Scan(&r.Day, &r.Turns); err != nil {
			return thermal.Summary{}, nil, err
		}
		r.Tokens = int64(r.Turns)
		daily = append(daily, r)
	}

	return summary, daily, nil
}
