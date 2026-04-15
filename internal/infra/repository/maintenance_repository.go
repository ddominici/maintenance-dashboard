package repository

import (
	"context"
	"fmt"

	"maintenance-dashboard/internal/domain/commandlog"
)

// dateGroupExpr returns a safe, pre-defined SQL expression that buckets
// StartTime into the requested granularity. Input is validated by the handler
// before it reaches here, so no injection risk.
func dateGroupExpr(granularity string) string {
	switch granularity {
	case "week":
		// Monday of the ISO week, independent of SET DATEFIRST.
		return "CONVERT(varchar(10), DATEADD(WEEK, DATEDIFF(WEEK, 0, StartTime), 0), 23)"
	case "month":
		return "LEFT(CONVERT(varchar(10), StartTime, 23), 7)"
	default: // "day"
		return "CONVERT(varchar(10), StartTime, 23)"
	}
}

func (r *CommandLogRepository) GetMaintenanceReport(ctx context.Context, filters commandlog.QueryFilters, granularity string) (commandlog.MaintenanceReport, error) {
	where, args := buildFilters(filters, 1)

	// ── Summary table ─────────────────────────────────────────────────────────
	summaryQuery := `
SELECT
    ISNULL(CommandType, '') AS CommandType,
    COUNT(*) AS TotalOperations,
    ISNULL(SUM(CASE WHEN ISNULL(ErrorNumber, 0)  = 0 THEN 1 ELSE 0 END), 0) AS Successes,
    ISNULL(SUM(CASE WHEN ISNULL(ErrorNumber, 0) <> 0 THEN 1 ELSE 0 END), 0) AS Errors,
    ISNULL(SUM(CAST(CASE WHEN EndTime IS NOT NULL THEN DATEDIFF(SECOND, StartTime, EndTime) ELSE 0 END AS bigint)), 0) AS TotalDurationSeconds,
    ISNULL(CAST(AVG(CASE WHEN EndTime IS NOT NULL THEN DATEDIFF(SECOND, StartTime, EndTime) * 1.0 END) AS float), 0.0) AS AvgDurationSeconds,
    MIN(StartTime) AS FirstOperation,
    MAX(StartTime) AS LastOperation
FROM dbo.CommandLog` + where + `
GROUP BY CommandType
ORDER BY TotalOperations DESC`

	sRows, err := r.db.QueryContext(ctx, summaryQuery, args...)
	if err != nil {
		return commandlog.MaintenanceReport{}, fmt.Errorf("query maintenance summary: %w", err)
	}
	var summary []commandlog.MaintenanceSummaryRow
	for sRows.Next() {
		var row commandlog.MaintenanceSummaryRow
		if err := sRows.Scan(
			&row.CommandType,
			&row.TotalOperations,
			&row.Successes,
			&row.ErrorCount,
			&row.TotalDurationSeconds,
			&row.AvgDurationSeconds,
			&row.FirstOperation,
			&row.LastOperation,
		); err != nil {
			sRows.Close()
			return commandlog.MaintenanceReport{}, fmt.Errorf("scan summary row: %w", err)
		}
		summary = append(summary, row)
	}
	if err := sRows.Close(); err != nil {
		return commandlog.MaintenanceReport{}, fmt.Errorf("close summary rows: %w", err)
	}
	if err := sRows.Err(); err != nil {
		return commandlog.MaintenanceReport{}, fmt.Errorf("iterate summary rows: %w", err)
	}

	// ── Timeline chart ────────────────────────────────────────────────────────
	dateExpr := dateGroupExpr(granularity)
	timelineQuery := fmt.Sprintf(`
SELECT
    %s AS OperationDate,
    ISNULL(CommandType, '') AS CommandType,
    ISNULL(SUM(CAST(CASE WHEN EndTime IS NOT NULL THEN DATEDIFF(SECOND, StartTime, EndTime) ELSE 0 END AS bigint)), 0) AS TotalDurationSeconds
FROM dbo.CommandLog%s
GROUP BY %s, CommandType
ORDER BY OperationDate, CommandType`, dateExpr, where, dateExpr)

	tRows, err := r.db.QueryContext(ctx, timelineQuery, args...)
	if err != nil {
		return commandlog.MaintenanceReport{}, fmt.Errorf("query maintenance timeline: %w", err)
	}
	var timeline []commandlog.TimelinePoint
	for tRows.Next() {
		var p commandlog.TimelinePoint
		if err := tRows.Scan(&p.Date, &p.CommandType, &p.TotalDurationSeconds); err != nil {
			tRows.Close()
			return commandlog.MaintenanceReport{}, fmt.Errorf("scan timeline row: %w", err)
		}
		timeline = append(timeline, p)
	}
	if err := tRows.Close(); err != nil {
		return commandlog.MaintenanceReport{}, fmt.Errorf("close timeline rows: %w", err)
	}
	if err := tRows.Err(); err != nil {
		return commandlog.MaintenanceReport{}, fmt.Errorf("iterate timeline rows: %w", err)
	}

	if summary == nil {
		summary = []commandlog.MaintenanceSummaryRow{}
	}
	if timeline == nil {
		timeline = []commandlog.TimelinePoint{}
	}
	return commandlog.MaintenanceReport{Summary: summary, Timeline: timeline}, nil
}
