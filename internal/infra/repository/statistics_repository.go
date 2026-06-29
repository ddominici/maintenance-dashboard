package repository

import (
	"context"
	"database/sql"
	"fmt"

	"maintenance-dashboard/internal/domain/commandlog"
)

func (r *CommandLogRepository) GetMostModifiedStatistics(ctx context.Context, filters commandlog.QueryFilters, limit int, sort commandlog.SortSpec) ([]commandlog.StatisticsRow, error) {
	// Force CommandType so callers don't need to set it.
	filters.CommandType = "UPDATE_STATISTICS"
	// @p1 is reserved for TOP; filter params start at @p2.
	where, filterArgs := buildFilters(filters, 2)

	// Whitelisted sort columns; default ranks by update count.
	orderBy := buildOrderBy(map[string]string{
		"updateCount":             "UpdateCount",
		"lastModificationCounter": "LastModificationCounter",
		"maxModificationCounter":  "MaxModificationCounter",
	}, sort, "UpdateCount DESC")

	// The CTE adds rn=1 to the most-recent row per object (by StartTime DESC).
	// This lets us read the *last* PageCount and ModificationCounter without a
	// correlated subquery, while MAX() covers the highest value ever seen.
	// ExtendedInfo is SQL Server XML; .value() returns NULL when the node is
	// absent or ExtendedInfo itself is NULL, so no extra NULL guard is needed.
	// The aggregated result is wrapped in a derived table so that the SELECT
	// aliases (UpdateCount, LastModificationCounter, MaxModificationCounter, ...)
	// become real columns. This lets buildOrderBy reference them inside
	// expressions (e.g. the NULLs-last CASE) — SQL Server rejects aliases used
	// within an ORDER BY expression, only allowing a bare alias reference.
	query := `
WITH cte AS (
    SELECT
        DatabaseName,
        SchemaName,
        ObjectName,
        StartTime,
        EndTime,
        ErrorNumber,
        ExtendedInfo,
        ROW_NUMBER() OVER (
            PARTITION BY DatabaseName, SchemaName, ObjectName
            ORDER BY StartTime DESC
        ) AS rn
    FROM dbo.CommandLog` + where + `
)
SELECT TOP (@p1) * FROM (
    SELECT
        ISNULL(DatabaseName, '') AS DatabaseName,
        ISNULL(SchemaName,  '') AS SchemaName,
        ISNULL(ObjectName,  '') AS ObjectName,
        COUNT(*) AS UpdateCount,
        MAX(StartTime) AS LastUpdated,
        ISNULL(CAST(AVG(CASE WHEN EndTime IS NOT NULL THEN DATEDIFF(SECOND, StartTime, EndTime) * 1.0 END) AS float), 0.0) AS AvgDurationSeconds,
        ISNULL(SUM(CASE WHEN ISNULL(ErrorNumber, 0) <> 0 THEN 1 ELSE 0 END), 0) AS ErrorCount,
        MAX(CASE WHEN rn = 1 THEN ExtendedInfo.value('(/ExtendedInfo/PageCount)[1]',           'bigint') END) AS LastPageCount,
        MAX(CASE WHEN rn = 1 THEN ExtendedInfo.value('(/ExtendedInfo/ModificationCounter)[1]', 'bigint') END) AS LastModificationCounter,
        MAX(ExtendedInfo.value('(/ExtendedInfo/ModificationCounter)[1]', 'bigint'))                           AS MaxModificationCounter
    FROM cte
    GROUP BY DatabaseName, SchemaName, ObjectName
) AS agg` + orderBy

	args := append([]any{limit}, filterArgs...)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query most modified statistics: %w", err)
	}

	var out []commandlog.StatisticsRow
	for rows.Next() {
		var row commandlog.StatisticsRow
		var lastPageCount, lastModCounter, maxModCounter sql.NullInt64
		if err := rows.Scan(
			&row.Database,
			&row.Schema,
			&row.Object,
			&row.UpdateCount,
			&row.LastUpdated,
			&row.AvgDurationSeconds,
			&row.ErrorCount,
			&lastPageCount,
			&lastModCounter,
			&maxModCounter,
		); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan statistics row: %w", err)
		}
		if lastPageCount.Valid {
			row.LastPageCount = &lastPageCount.Int64
		}
		if lastModCounter.Valid {
			row.LastModificationCounter = &lastModCounter.Int64
		}
		if maxModCounter.Valid {
			row.MaxModificationCounter = &maxModCounter.Int64
		}
		out = append(out, row)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close statistics rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate statistics rows: %w", err)
	}

	if out == nil {
		out = []commandlog.StatisticsRow{}
	}
	return out, nil
}
