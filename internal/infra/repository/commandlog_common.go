package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"maintenance-dashboard/internal/domain/commandlog"
)

type CommandLogRepository struct {
	db *sql.DB
}

func NewCommandLogRepository(db *sql.DB) *CommandLogRepository {
	return &CommandLogRepository{db: db}
}

func buildFilters(filters commandlog.QueryFilters, startIndex int) (string, []any) {
	clauses := make([]string, 0, 6)
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
	if filters.OnlyErrors {
		clauses = append(clauses, "ISNULL(ErrorNumber, 0) <> 0")
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

// buildOrderBy produces an ORDER BY clause from a caller-supplied SortSpec.
// columns whitelists the logical sort keys → SQL column/alias names, so the
// final SQL never contains raw user input (guarding against injection). When
// sort.By is empty or unknown, the query's default ordering (def, e.g.
// "TotalOperations DESC") is used. NULLs are always ordered last regardless of
// direction, matching the previous client-side sort behavior.
func buildOrderBy(columns map[string]string, sort commandlog.SortSpec, def string) string {
	col, ok := columns[sort.By]
	if !ok {
		return " ORDER BY " + def
	}
	dir := "DESC"
	if strings.EqualFold(sort.Dir, "asc") {
		dir = "ASC"
	}
	return fmt.Sprintf(" ORDER BY CASE WHEN %s IS NULL THEN 1 ELSE 0 END, %s %s", col, col, dir)
}
