package repository

import (
	"context"
	"fmt"

	"maintenance-dashboard/internal/domain/commandlog"
)

func (r *CommandLogRepository) GetOperationsBatchReport(ctx context.Context, filters commandlog.QueryFilters, granularity string) (commandlog.OperationsBatchReport, error) {
	where, args := buildFilters(filters, 1)

	// ── Summary table ─────────────────────────────────────────────────────────
	summaryQuery := `
SELECT
    ISNULL(CommandType, '') AS CommandType,
    COUNT(*) AS TotalCount,
    ISNULL(SUM(CASE WHEN ISNULL(ErrorNumber, 0)  = 0 THEN 1 ELSE 0 END), 0) AS SuccessCount,
    ISNULL(SUM(CASE WHEN ISNULL(ErrorNumber, 0) <> 0 THEN 1 ELSE 0 END), 0) AS ErrorCount,
    MIN(StartTime) AS FirstOperation,
    MAX(StartTime) AS LastOperation
FROM dbo.CommandLog` + where + `
GROUP BY CommandType
ORDER BY TotalCount DESC`

	sRows, err := r.db.QueryContext(ctx, summaryQuery, args...)
	if err != nil {
		return commandlog.OperationsBatchReport{}, fmt.Errorf("query operations summary: %w", err)
	}
	var summary []commandlog.OperationsSummaryRow
	for sRows.Next() {
		var row commandlog.OperationsSummaryRow
		if err := sRows.Scan(
			&row.CommandType,
			&row.TotalCount,
			&row.SuccessCount,
			&row.ErrorCount,
			&row.FirstOperation,
			&row.LastOperation,
		); err != nil {
			sRows.Close()
			return commandlog.OperationsBatchReport{}, fmt.Errorf("scan operations summary row: %w", err)
		}
		summary = append(summary, row)
	}
	if err := sRows.Close(); err != nil {
		return commandlog.OperationsBatchReport{}, fmt.Errorf("close operations summary rows: %w", err)
	}
	if err := sRows.Err(); err != nil {
		return commandlog.OperationsBatchReport{}, fmt.Errorf("iterate operations summary rows: %w", err)
	}

	// ── Timeline chart ────────────────────────────────────────────────────────
	dateExpr := dateGroupExpr(granularity)
	timelineQuery := fmt.Sprintf(`
SELECT
    %s AS OperationDate,
    ISNULL(CommandType, '') AS CommandType,
    COUNT(*) AS OperationCount,
    ISNULL(SUM(CASE WHEN ISNULL(ErrorNumber, 0) <> 0 THEN 1 ELSE 0 END), 0) AS ErrorCount
FROM dbo.CommandLog%s
GROUP BY %s, CommandType
ORDER BY OperationDate, CommandType`, dateExpr, where, dateExpr)

	tRows, err := r.db.QueryContext(ctx, timelineQuery, args...)
	if err != nil {
		return commandlog.OperationsBatchReport{}, fmt.Errorf("query operations timeline: %w", err)
	}
	var timeline []commandlog.OperationsBatchPoint
	for tRows.Next() {
		var p commandlog.OperationsBatchPoint
		if err := tRows.Scan(&p.Date, &p.CommandType, &p.Count, &p.ErrorCount); err != nil {
			tRows.Close()
			return commandlog.OperationsBatchReport{}, fmt.Errorf("scan operations timeline row: %w", err)
		}
		timeline = append(timeline, p)
	}
	if err := tRows.Close(); err != nil {
		return commandlog.OperationsBatchReport{}, fmt.Errorf("close operations timeline rows: %w", err)
	}
	if err := tRows.Err(); err != nil {
		return commandlog.OperationsBatchReport{}, fmt.Errorf("iterate operations timeline rows: %w", err)
	}

	if summary == nil {
		summary = []commandlog.OperationsSummaryRow{}
	}
	if timeline == nil {
		timeline = []commandlog.OperationsBatchPoint{}
	}
	return commandlog.OperationsBatchReport{Summary: summary, Timeline: timeline}, nil
}
