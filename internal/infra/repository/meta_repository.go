package repository

import (
	"context"
	"fmt"

	"maintenance-dashboard/internal/domain/commandlog"
)

func (r *CommandLogRepository) GetFilterOptions(ctx context.Context) (commandlog.FilterOptions, error) {
	out := commandlog.FilterOptions{}

	q1 := `SELECT DISTINCT DatabaseName FROM dbo.CommandLog WHERE DatabaseName IS NOT NULL ORDER BY DatabaseName`
	rows, err := r.db.QueryContext(ctx, q1)
	if err != nil {
		return out, fmt.Errorf("query databases: %w", err)
	}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			rows.Close()
			return out, fmt.Errorf("scan database: %w", err)
		}
		out.Databases = append(out.Databases, s)
	}
	if err := rows.Close(); err != nil {
		return out, fmt.Errorf("close databases rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return out, fmt.Errorf("iterate databases: %w", err)
	}

	q2 := `SELECT DISTINCT SchemaName FROM dbo.CommandLog WHERE SchemaName IS NOT NULL ORDER BY SchemaName`
	rows, err = r.db.QueryContext(ctx, q2)
	if err != nil {
		return out, fmt.Errorf("query schemas: %w", err)
	}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			rows.Close()
			return out, fmt.Errorf("scan schema: %w", err)
		}
		out.Schemas = append(out.Schemas, s)
	}
	if err := rows.Close(); err != nil {
		return out, fmt.Errorf("close schemas rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return out, fmt.Errorf("iterate schemas: %w", err)
	}

	q3 := `SELECT DISTINCT CommandType FROM dbo.CommandLog WHERE CommandType IS NOT NULL ORDER BY CommandType`
	rows, err = r.db.QueryContext(ctx, q3)
	if err != nil {
		return out, fmt.Errorf("query command types: %w", err)
	}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			rows.Close()
			return out, fmt.Errorf("scan command type: %w", err)
		}
		out.CommandTypes = append(out.CommandTypes, s)
	}
	if err := rows.Close(); err != nil {
		return out, fmt.Errorf("close command types rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return out, fmt.Errorf("iterate command types: %w", err)
	}

	return out, nil
}
