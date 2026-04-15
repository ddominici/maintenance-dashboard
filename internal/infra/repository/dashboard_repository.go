package repository

import (
	"context"
	"fmt"

	"maintenance-dashboard/internal/domain/commandlog"
)

func (r *CommandLogRepository) GetSummary(ctx context.Context, filters commandlog.QueryFilters) (commandlog.DashboardSummary, error) {
	where, args := buildFilters(filters, 1)
	query := `
SELECT
    COUNT_BIG(*) AS TotalCommands,
    ISNULL(SUM(CASE WHEN ISNULL(ErrorNumber, 0) <> 0 THEN 1 ELSE 0 END), 0) AS TotalErrors,
    ISNULL(CAST(AVG(CASE WHEN EndTime IS NOT NULL THEN DATEDIFF(SECOND, StartTime, EndTime) * 1.0 END) AS float), 0.0) AS AvgDurationSeconds,
    ISNULL(SUM(CASE WHEN CommandType = 'ALTER_INDEX' AND Command LIKE '%REBUILD%' THEN 1 ELSE 0 END), 0) AS TotalIndexRebuilds,
    ISNULL(SUM(CASE WHEN CommandType = 'ALTER_INDEX' AND Command LIKE '%REORGANIZE%' THEN 1 ELSE 0 END), 0) AS TotalIndexReorganizes,
    ISNULL(SUM(CASE WHEN CommandType = 'UPDATE_STATISTICS' THEN 1 ELSE 0 END), 0) AS TotalStatisticsUpdates
FROM dbo.CommandLog` + where

	var out commandlog.DashboardSummary
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&out.TotalCommands,
		&out.TotalErrors,
		&out.AvgDurationSeconds,
		&out.TotalIndexRebuilds,
		&out.TotalIndexReorganizes,
		&out.TotalStatisticsUpdates,
	); err != nil {
		return out, fmt.Errorf("query dashboard summary: %w", err)
	}
	return out, nil
}
