package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"maintenance-dashboard/internal/domain/commandlog"
)

// buildLongRunningFilters always prepends the EndTime IS NOT NULL and minimum
// duration conditions before any caller-supplied filters. The minDurationSeconds
// value occupies parameter @p<startIndex>; subsequent filter params follow.
func buildLongRunningFilters(filters commandlog.QueryFilters, minDurationSeconds int, startIndex int) (string, []any) {
	clauses := []string{
		"EndTime IS NOT NULL",
		fmt.Sprintf("DATEDIFF(SECOND, StartTime, EndTime) >= @p%d", startIndex),
	}
	args := []any{minDurationSeconds}
	idx := startIndex + 1
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
	if filters.OnlyErrors {
		clauses = append(clauses, "ISNULL(ErrorNumber, 0) <> 0")
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *CommandLogRepository) GetLongRunningReport(
	ctx context.Context,
	filters commandlog.QueryFilters,
	granularity string,
	minDurationSeconds int,
	limit int,
) (commandlog.LongRunningReport, error) {

	// ── Timeline chart ────────────────────────────────────────────────────────
	// @p1 = minDurationSeconds; filter params start at @p2.
	timelineWhere, timelineArgs := buildLongRunningFilters(filters, minDurationSeconds, 1)
	dateExpr := dateGroupExpr(granularity)

	timelineQuery := fmt.Sprintf(`
SELECT
    %s AS OperationDate,
    ISNULL(CommandType, '') AS CommandType,
    MAX(DATEDIFF(SECOND, StartTime, EndTime)) AS MaxDurationSeconds,
    CAST(AVG(DATEDIFF(SECOND, StartTime, EndTime) * 1.0) AS float) AS AvgDurationSeconds,
    COUNT(*) AS OperationCount
FROM dbo.CommandLog%s
GROUP BY %s, CommandType
ORDER BY OperationDate, CommandType`, dateExpr, timelineWhere, dateExpr)

	tRows, err := r.db.QueryContext(ctx, timelineQuery, timelineArgs...)
	if err != nil {
		return commandlog.LongRunningReport{}, fmt.Errorf("query long running timeline: %w", err)
	}
	var timeline []commandlog.LongRunningTimelinePoint
	for tRows.Next() {
		var p commandlog.LongRunningTimelinePoint
		if err := tRows.Scan(&p.Date, &p.CommandType, &p.MaxDurationSeconds, &p.AvgDurationSeconds, &p.Count); err != nil {
			tRows.Close()
			return commandlog.LongRunningReport{}, fmt.Errorf("scan long running timeline row: %w", err)
		}
		timeline = append(timeline, p)
	}
	if err := tRows.Close(); err != nil {
		return commandlog.LongRunningReport{}, fmt.Errorf("close long running timeline rows: %w", err)
	}
	if err := tRows.Err(); err != nil {
		return commandlog.LongRunningReport{}, fmt.Errorf("iterate long running timeline rows: %w", err)
	}

	// ── Top operations table ──────────────────────────────────────────────────
	// @p1 = limit; @p2 = minDurationSeconds; filter params start at @p3.
	opsWhere, opsFilterArgs := buildLongRunningFilters(filters, minDurationSeconds, 2)
	opsArgs := append([]any{limit}, opsFilterArgs...)

	opsQuery := `
SELECT TOP (@p1)
    ISNULL(DatabaseName, '') AS DatabaseName,
    ISNULL(SchemaName,  '') AS SchemaName,
    ISNULL(ObjectName,  '') AS ObjectName,
    ISNULL(IndexName,   '') AS IndexName,
    ISNULL(CommandType, '') AS CommandType,
    StartTime,
    DATEDIFF(SECOND, StartTime, EndTime) AS DurationSeconds,
    ErrorNumber,
    ErrorMessage
FROM dbo.CommandLog` + opsWhere + `
ORDER BY DurationSeconds DESC`

	oRows, err := r.db.QueryContext(ctx, opsQuery, opsArgs...)
	if err != nil {
		return commandlog.LongRunningReport{}, fmt.Errorf("query long running operations: %w", err)
	}
	var operations []commandlog.LongRunningRow
	for oRows.Next() {
		var row commandlog.LongRunningRow
		var errNum sql.NullInt32
		var errMsg sql.NullString
		if err := oRows.Scan(
			&row.Database,
			&row.Schema,
			&row.Object,
			&row.Index,
			&row.CommandType,
			&row.StartTime,
			&row.DurationSeconds,
			&errNum,
			&errMsg,
		); err != nil {
			oRows.Close()
			return commandlog.LongRunningReport{}, fmt.Errorf("scan long running operation row: %w", err)
		}
		if errNum.Valid {
			row.ErrorNumber = &errNum.Int32
		}
		if errMsg.Valid {
			row.ErrorMessage = &errMsg.String
		}
		operations = append(operations, row)
	}
	if err := oRows.Close(); err != nil {
		return commandlog.LongRunningReport{}, fmt.Errorf("close long running operation rows: %w", err)
	}
	if err := oRows.Err(); err != nil {
		return commandlog.LongRunningReport{}, fmt.Errorf("iterate long running operation rows: %w", err)
	}

	if timeline == nil {
		timeline = []commandlog.LongRunningTimelinePoint{}
	}
	if operations == nil {
		operations = []commandlog.LongRunningRow{}
	}
	return commandlog.LongRunningReport{Operations: operations, Timeline: timeline}, nil
}
