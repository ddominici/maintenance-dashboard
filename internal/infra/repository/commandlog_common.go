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
