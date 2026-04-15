package commandlog

import (
	"context"
	"time"
)

type FilterOptions struct {
	Databases    []string `json:"databases"`
	Schemas      []string `json:"schemas"`
	CommandTypes []string `json:"commandTypes"`
}

type DashboardSummary struct {
	TotalCommands          int64   `json:"totalCommands"`
	TotalErrors            int64   `json:"totalErrors"`
	AvgDurationSeconds     float64 `json:"avgDurationSeconds"`
	TotalIndexRebuilds     int64   `json:"totalIndexRebuilds"`
	TotalIndexReorganizes  int64   `json:"totalIndexReorganizes"`
	TotalStatisticsUpdates int64   `json:"totalStatisticsUpdates"`
}

type StatisticsRow struct {
	Database           string    `json:"database"`
	Schema             string    `json:"schema"`
	Object             string    `json:"object"`
	UpdateCount        int64     `json:"updateCount"`
	LastUpdated        time.Time `json:"lastUpdated"`
	AvgDurationSeconds float64   `json:"avgDurationSeconds"`
	ErrorCount         int64     `json:"errorCount"`
	// Extracted from ExtendedInfo XML; nil when the column is absent or NULL.
	LastPageCount           *int64 `json:"lastPageCount"`
	LastModificationCounter *int64 `json:"lastModificationCounter"`
	MaxModificationCounter  *int64 `json:"maxModificationCounter"`
}

type MetaRepository interface {
	GetFilterOptions(ctx context.Context) (FilterOptions, error)
}

type DashboardRepository interface {
	GetSummary(ctx context.Context, filters QueryFilters) (DashboardSummary, error)
}

type StatisticsRepository interface {
	GetMostModifiedStatistics(ctx context.Context, filters QueryFilters, limit int) ([]StatisticsRow, error)
}

type IndexRow struct {
	Database           string    `json:"database"`
	Schema             string    `json:"schema"`
	Object             string    `json:"object"`
	Index              string    `json:"index"`
	RebuildCount       int64     `json:"rebuildCount"`
	ReorganizeCount    int64     `json:"reorganizeCount"`
	TotalOperations    int64     `json:"totalOperations"`
	LastOperation      time.Time `json:"lastOperation"`
	AvgDurationSeconds float64   `json:"avgDurationSeconds"`
	ErrorCount         int64     `json:"errorCount"`
	// Extracted from ExtendedInfo XML; nil when the column is absent or NULL.
	MaxFragmentation  *float64 `json:"maxFragmentation"`
	LastFragmentation *float64 `json:"lastFragmentation"`
}

type IndexRepository interface {
	GetTopFragmentedIndexes(ctx context.Context, filters QueryFilters, limit int) ([]IndexRow, error)
}

type MaintenanceSummaryRow struct {
	CommandType          string    `json:"commandType"`
	TotalOperations      int64     `json:"totalOperations"`
	Successes            int64     `json:"successes"`
	ErrorCount           int64     `json:"errorCount"`
	TotalDurationSeconds int64     `json:"totalDurationSeconds"`
	AvgDurationSeconds   float64   `json:"avgDurationSeconds"`
	FirstOperation       time.Time `json:"firstOperation"`
	LastOperation        time.Time `json:"lastOperation"`
}

type TimelinePoint struct {
	Date                 string `json:"date"`
	CommandType          string `json:"commandType"`
	TotalDurationSeconds int64  `json:"totalDurationSeconds"`
}

type MaintenanceReport struct {
	Summary  []MaintenanceSummaryRow `json:"summary"`
	Timeline []TimelinePoint         `json:"timeline"`
}

type MaintenanceRepository interface {
	GetMaintenanceReport(ctx context.Context, filters QueryFilters, granularity string) (MaintenanceReport, error)
}

type ErrorSummaryRow struct {
	CommandType string    `json:"commandType"`
	ErrorCount  int64     `json:"errorCount"`
	FirstError  time.Time `json:"firstError"`
	LastError   time.Time `json:"lastError"`
}

type ErrorTimelinePoint struct {
	Date        string `json:"date"`
	CommandType string `json:"commandType"`
	ErrorCount  int64  `json:"errorCount"`
}

type ErrorDetailRow struct {
	Database        string    `json:"database"`
	Schema          string    `json:"schema"`
	Object          string    `json:"object"`
	Index           string    `json:"index"`
	CommandType     string    `json:"commandType"`
	StartTime       time.Time `json:"startTime"`
	DurationSeconds *int64    `json:"durationSeconds"` // nil when EndTime is NULL
	ErrorNumber     int32     `json:"errorNumber"`
	ErrorMessage    *string   `json:"errorMessage"`
}

type MaintenanceErrorsReport struct {
	Summary  []ErrorSummaryRow    `json:"summary"`
	Timeline []ErrorTimelinePoint `json:"timeline"`
	Detail   []ErrorDetailRow     `json:"detail"`
}

type MaintenanceErrorsRepository interface {
	GetMaintenanceErrorsReport(ctx context.Context, filters QueryFilters, granularity string, limit int) (MaintenanceErrorsReport, error)
}

type LongRunningRow struct {
	Database        string     `json:"database"`
	Schema          string     `json:"schema"`
	Object          string     `json:"object"`
	Index           string     `json:"index"`
	CommandType     string     `json:"commandType"`
	StartTime       time.Time  `json:"startTime"`
	DurationSeconds int64      `json:"durationSeconds"`
	ErrorNumber     *int32     `json:"errorNumber"`
	ErrorMessage    *string    `json:"errorMessage"`
}

type LongRunningTimelinePoint struct {
	Date               string  `json:"date"`
	CommandType        string  `json:"commandType"`
	MaxDurationSeconds int64   `json:"maxDurationSeconds"`
	AvgDurationSeconds float64 `json:"avgDurationSeconds"`
	Count              int64   `json:"count"`
}

type LongRunningReport struct {
	Operations []LongRunningRow           `json:"operations"`
	Timeline   []LongRunningTimelinePoint `json:"timeline"`
}

type LongRunningRepository interface {
	GetLongRunningReport(ctx context.Context, filters QueryFilters, granularity string, minDurationSeconds int, limit int) (LongRunningReport, error)
}

type OperationsBatchPoint struct {
	Date        string `json:"date"`
	CommandType string `json:"commandType"`
	Count       int64  `json:"count"`
	ErrorCount  int64  `json:"errorCount"`
}

type OperationsSummaryRow struct {
	CommandType    string    `json:"commandType"`
	TotalCount     int64     `json:"totalCount"`
	SuccessCount   int64     `json:"successCount"`
	ErrorCount     int64     `json:"errorCount"`
	FirstOperation time.Time `json:"firstOperation"`
	LastOperation  time.Time `json:"lastOperation"`
}

type OperationsBatchReport struct {
	Summary  []OperationsSummaryRow `json:"summary"`
	Timeline []OperationsBatchPoint `json:"timeline"`
}

type OperationsBatchRepository interface {
	GetOperationsBatchReport(ctx context.Context, filters QueryFilters, granularity string) (OperationsBatchReport, error)
}

type BackupTimelinePoint struct {
	Date                 string  `json:"date"`
	Database             string  `json:"database"`
	TotalCount           int64   `json:"totalCount"`
	SuccessCount         int64   `json:"successCount"`
	ErrorCount           int64   `json:"errorCount"`
	TotalDurationSeconds int64   `json:"totalDurationSeconds"`
	AvgDurationSeconds   float64 `json:"avgDurationSeconds"`
}

type BackupReport struct {
	Timeline []BackupTimelinePoint `json:"timeline"`
}

type BackupRepository interface {
	GetBackupReport(ctx context.Context, filters QueryFilters, granularity string) (BackupReport, error)
}
