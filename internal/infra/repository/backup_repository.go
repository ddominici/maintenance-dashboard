package repository

import (
	"context"
	"fmt"
	"strings"

	"maintenance-dashboard/internal/domain/commandlog"
)

// buildBackupFilters is like buildFilters but always adds the CommandType LIKE 'BACKUP%'
// condition as the first clause, ignoring any CommandType value in the filters struct.
func buildBackupFilters(filters commandlog.QueryFilters, startIndex int) (string, []any) {
	clauses := []string{"CommandType LIKE 'BACKUP%'"}
	args := make([]any, 0, 6)
	idx := startIndex
	add := func(clause string, val any) {
		clauses = append(clauses, fmt.Sprintf(clause, idx))
		args = append(args, val)
		idx++
	}
	if filters.DateFrom != nil {
		add("StartTime >= @p%d", *filters.DateFrom)
	}
	if filters.DateTo != nil {
		add("StartTime < @p%d", *filters.DateTo)
	}
	if filters.Database != "" {
		add("DatabaseName = @p%d", filters.Database)
	}
	if filters.Schema != "" {
		add("SchemaName = @p%d", filters.Schema)
	}
	if filters.Object != "" {
		add("ObjectName = @p%d", filters.Object)
	}
	if filters.OnlyErrors {
		clauses = append(clauses, "ISNULL(ErrorNumber, 0) <> 0")
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *CommandLogRepository) GetBackupReport(ctx context.Context, filters commandlog.QueryFilters, granularity string) (commandlog.BackupReport, error) {
	where, args := buildBackupFilters(filters, 1)
	dateExpr := dateGroupExpr(granularity)

	query := fmt.Sprintf(`
SELECT
    %s AS OperationDate,
    ISNULL(DatabaseName, '') AS DatabaseName,
    COUNT(*) AS TotalCount,
    ISNULL(SUM(CASE WHEN ISNULL(ErrorNumber, 0)  = 0 THEN 1 ELSE 0 END), 0) AS SuccessCount,
    ISNULL(SUM(CASE WHEN ISNULL(ErrorNumber, 0) <> 0 THEN 1 ELSE 0 END), 0) AS ErrorCount,
    ISNULL(SUM(CAST(CASE WHEN EndTime IS NOT NULL THEN DATEDIFF(SECOND, StartTime, EndTime) ELSE 0 END AS bigint)), 0) AS TotalDurationSeconds,
    ISNULL(CAST(AVG(CASE WHEN EndTime IS NOT NULL AND ISNULL(ErrorNumber, 0) = 0 THEN DATEDIFF(SECOND, StartTime, EndTime) * 1.0 END) AS float), 0.0) AS AvgDurationSeconds
FROM dbo.CommandLog%s
GROUP BY %s, ISNULL(DatabaseName, '')
ORDER BY OperationDate, DatabaseName`, dateExpr, where, dateExpr)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return commandlog.BackupReport{}, fmt.Errorf("query backup report: %w", err)
	}

	var timeline []commandlog.BackupTimelinePoint
	for rows.Next() {
		var p commandlog.BackupTimelinePoint
		if err := rows.Scan(
			&p.Date,
			&p.Database,
			&p.TotalCount,
			&p.SuccessCount,
			&p.ErrorCount,
			&p.TotalDurationSeconds,
			&p.AvgDurationSeconds,
		); err != nil {
			rows.Close()
			return commandlog.BackupReport{}, fmt.Errorf("scan backup row: %w", err)
		}
		timeline = append(timeline, p)
	}
	if err := rows.Close(); err != nil {
		return commandlog.BackupReport{}, fmt.Errorf("close backup rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return commandlog.BackupReport{}, fmt.Errorf("iterate backup rows: %w", err)
	}

	if timeline == nil {
		timeline = []commandlog.BackupTimelinePoint{}
	}
	return commandlog.BackupReport{Timeline: timeline}, nil
}
