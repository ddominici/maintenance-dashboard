package repository

import (
	"context"
	"database/sql"
	"fmt"

	"maintenance-dashboard/internal/domain/commandlog"
)

func (r *CommandLogRepository) GetTopFragmentedIndexes(ctx context.Context, filters commandlog.QueryFilters, limit int, sort commandlog.SortSpec) ([]commandlog.IndexRow, error) {
	filters.CommandType = "ALTER_INDEX"
	// @p1 is TOP limit; filter params start at @p2.
	where, filterArgs := buildFilters(filters, 2)

	// Whitelisted sort columns; default ranks by total maintenance operations.
	orderBy := buildOrderBy(map[string]string{
		"totalOperations":   "TotalOperations",
		"lastFragmentation": "LastFragmentation",
		"maxFragmentation":  "MaxFragmentation",
	}, sort, "TotalOperations DESC")

	// The CTE adds rn=1 to the most-recent row per index (by StartTime DESC).
	// This lets us extract the *last* fragmentation value without a correlated
	// subquery, while MAX() in the outer SELECT gives the *highest* ever seen.
	// ExtendedInfo is SQL Server XML; .value() returns NULL when the node is
	// absent or ExtendedInfo itself is NULL, so no extra NULL guard is needed.
	// The aggregated result is wrapped in a derived table so that the SELECT
	// aliases (TotalOperations, MaxFragmentation, LastFragmentation, ...) become
	// real columns. This lets buildOrderBy reference them inside expressions
	// (e.g. the NULLs-last CASE) — SQL Server rejects aliases used within an
	// ORDER BY expression, only allowing a bare alias reference.
	query := `
WITH cte AS (
    SELECT
        DatabaseName,
        SchemaName,
        ObjectName,
        IndexName,
        Command,
        StartTime,
        EndTime,
        ErrorNumber,
        ExtendedInfo,
        ROW_NUMBER() OVER (
            PARTITION BY DatabaseName, SchemaName, ObjectName, IndexName
            ORDER BY StartTime DESC
        ) AS rn
    FROM dbo.CommandLog` + where + `
)
SELECT TOP (@p1) * FROM (
    SELECT
        ISNULL(DatabaseName, '') AS DatabaseName,
        ISNULL(SchemaName,  '') AS SchemaName,
        ISNULL(ObjectName,  '') AS ObjectName,
        ISNULL(IndexName,   '') AS IndexName,
        ISNULL(SUM(CASE WHEN Command LIKE '%REBUILD%'    THEN 1 ELSE 0 END), 0) AS RebuildCount,
        ISNULL(SUM(CASE WHEN Command LIKE '%REORGANIZE%' THEN 1 ELSE 0 END), 0) AS ReorganizeCount,
        COUNT(*) AS TotalOperations,
        MAX(StartTime) AS LastOperation,
        ISNULL(CAST(AVG(CASE WHEN EndTime IS NOT NULL THEN DATEDIFF(SECOND, StartTime, EndTime) * 1.0 END) AS float), 0.0) AS AvgDurationSeconds,
        ISNULL(SUM(CASE WHEN ISNULL(ErrorNumber, 0) <> 0 THEN 1 ELSE 0 END), 0) AS ErrorCount,
        MAX(ExtendedInfo.value('(/ExtendedInfo/Fragmentation)[1]', 'float'))                             AS MaxFragmentation,
        MAX(CASE WHEN rn = 1 THEN ExtendedInfo.value('(/ExtendedInfo/Fragmentation)[1]', 'float') END)  AS LastFragmentation
    FROM cte
    GROUP BY DatabaseName, SchemaName, ObjectName, IndexName
) AS agg` + orderBy

	args := append([]any{limit}, filterArgs...)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query top fragmented indexes: %w", err)
	}

	var out []commandlog.IndexRow
	for rows.Next() {
		var row commandlog.IndexRow
		// MaxFragmentation and LastFragmentation are nullable: use NullFloat64
		// so that absent/NULL XML nodes map cleanly to a nil *float64.
		var maxFrag, lastFrag sql.NullFloat64
		if err := rows.Scan(
			&row.Database,
			&row.Schema,
			&row.Object,
			&row.Index,
			&row.RebuildCount,
			&row.ReorganizeCount,
			&row.TotalOperations,
			&row.LastOperation,
			&row.AvgDurationSeconds,
			&row.ErrorCount,
			&maxFrag,
			&lastFrag,
		); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan index row: %w", err)
		}
		if maxFrag.Valid {
			row.MaxFragmentation = &maxFrag.Float64
		}
		if lastFrag.Valid {
			row.LastFragmentation = &lastFrag.Float64
		}
		out = append(out, row)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close index rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate index rows: %w", err)
	}

	if out == nil {
		out = []commandlog.IndexRow{}
	}
	return out, nil
}
