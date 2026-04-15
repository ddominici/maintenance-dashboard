package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"maintenance-dashboard/internal/domain/commandlog"
)

// buildErrorFilters always prepends the ISNULL(ErrorNumber,0) <> 0 condition
// before any caller-supplied filters. OnlyErrors in the filters struct is ignored
// because it is implied by this endpoint.
func buildErrorFilters(filters commandlog.QueryFilters, startIndex int) (string, []any) {
	clauses := []string{"ISNULL(ErrorNumber, 0) <> 0"}
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
	if filters.CommandType != "" {
		add("CommandType = @p%d", filters.CommandType)
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *CommandLogRepository) GetMaintenanceErrorsReport(
	ctx context.Context,
	filters commandlog.QueryFilters,
	granularity string,
	limit int,
) (commandlog.MaintenanceErrorsReport, error) {

	// ── Summary table ─────────────────────────────────────────────────────────
	// filter params start at @p1
	summaryWhere, summaryArgs := buildErrorFilters(filters, 1)

	summaryQuery := `
SELECT
    ISNULL(CommandType, '') AS CommandType,
    COUNT(*) AS ErrorCount,
    MIN(StartTime) AS FirstError,
    MAX(StartTime) AS LastError
FROM dbo.CommandLog` + summaryWhere + `
GROUP BY CommandType
ORDER BY ErrorCount DESC`

	sRows, err := r.db.QueryContext(ctx, summaryQuery, summaryArgs...)
	if err != nil {
		return commandlog.MaintenanceErrorsReport{}, fmt.Errorf("query errors summary: %w", err)
	}
	var summary []commandlog.ErrorSummaryRow
	for sRows.Next() {
		var row commandlog.ErrorSummaryRow
		if err := sRows.Scan(&row.CommandType, &row.ErrorCount, &row.FirstError, &row.LastError); err != nil {
			sRows.Close()
			return commandlog.MaintenanceErrorsReport{}, fmt.Errorf("scan errors summary row: %w", err)
		}
		summary = append(summary, row)
	}
	if err := sRows.Close(); err != nil {
		return commandlog.MaintenanceErrorsReport{}, fmt.Errorf("close errors summary rows: %w", err)
	}
	if err := sRows.Err(); err != nil {
		return commandlog.MaintenanceErrorsReport{}, fmt.Errorf("iterate errors summary rows: %w", err)
	}

	// ── Timeline chart ────────────────────────────────────────────────────────
	// Reuse same args (same filters, same param positions).
	dateExpr := dateGroupExpr(granularity)
	timelineQuery := fmt.Sprintf(`
SELECT
    %s AS OperationDate,
    ISNULL(CommandType, '') AS CommandType,
    COUNT(*) AS ErrorCount
FROM dbo.CommandLog%s
GROUP BY %s, CommandType
ORDER BY OperationDate, CommandType`, dateExpr, summaryWhere, dateExpr)

	tRows, err := r.db.QueryContext(ctx, timelineQuery, summaryArgs...)
	if err != nil {
		return commandlog.MaintenanceErrorsReport{}, fmt.Errorf("query errors timeline: %w", err)
	}
	var timeline []commandlog.ErrorTimelinePoint
	for tRows.Next() {
		var p commandlog.ErrorTimelinePoint
		if err := tRows.Scan(&p.Date, &p.CommandType, &p.ErrorCount); err != nil {
			tRows.Close()
			return commandlog.MaintenanceErrorsReport{}, fmt.Errorf("scan errors timeline row: %w", err)
		}
		timeline = append(timeline, p)
	}
	if err := tRows.Close(); err != nil {
		return commandlog.MaintenanceErrorsReport{}, fmt.Errorf("close errors timeline rows: %w", err)
	}
	if err := tRows.Err(); err != nil {
		return commandlog.MaintenanceErrorsReport{}, fmt.Errorf("iterate errors timeline rows: %w", err)
	}

	// ── Detail table ──────────────────────────────────────────────────────────
	// @p1 = limit; filter params start at @p2.
	detailWhere, detailFilterArgs := buildErrorFilters(filters, 2)
	detailArgs := append([]any{limit}, detailFilterArgs...)

	detailQuery := `
SELECT TOP (@p1)
    ISNULL(DatabaseName, '') AS DatabaseName,
    ISNULL(SchemaName,  '') AS SchemaName,
    ISNULL(ObjectName,  '') AS ObjectName,
    ISNULL(IndexName,   '') AS IndexName,
    ISNULL(CommandType, '') AS CommandType,
    StartTime,
    CASE WHEN EndTime IS NOT NULL THEN DATEDIFF(SECOND, StartTime, EndTime) END AS DurationSeconds,
    ISNULL(ErrorNumber, 0) AS ErrorNumber,
    ErrorMessage
FROM dbo.CommandLog` + detailWhere + `
ORDER BY StartTime DESC`

	dRows, err := r.db.QueryContext(ctx, detailQuery, detailArgs...)
	if err != nil {
		return commandlog.MaintenanceErrorsReport{}, fmt.Errorf("query errors detail: %w", err)
	}
	var detail []commandlog.ErrorDetailRow
	for dRows.Next() {
		var row commandlog.ErrorDetailRow
		var dur sql.NullInt64
		var msg sql.NullString
		if err := dRows.Scan(
			&row.Database,
			&row.Schema,
			&row.Object,
			&row.Index,
			&row.CommandType,
			&row.StartTime,
			&dur,
			&row.ErrorNumber,
			&msg,
		); err != nil {
			dRows.Close()
			return commandlog.MaintenanceErrorsReport{}, fmt.Errorf("scan errors detail row: %w", err)
		}
		if dur.Valid {
			row.DurationSeconds = &dur.Int64
		}
		if msg.Valid {
			row.ErrorMessage = &msg.String
		}
		detail = append(detail, row)
	}
	if err := dRows.Close(); err != nil {
		return commandlog.MaintenanceErrorsReport{}, fmt.Errorf("close errors detail rows: %w", err)
	}
	if err := dRows.Err(); err != nil {
		return commandlog.MaintenanceErrorsReport{}, fmt.Errorf("iterate errors detail rows: %w", err)
	}

	if summary == nil {
		summary = []commandlog.ErrorSummaryRow{}
	}
	if timeline == nil {
		timeline = []commandlog.ErrorTimelinePoint{}
	}
	if detail == nil {
		detail = []commandlog.ErrorDetailRow{}
	}
	return commandlog.MaintenanceErrorsReport{Summary: summary, Timeline: timeline, Detail: detail}, nil
}
