export interface ApiEnvelope<T> {
  data: T
}

export interface FilterOptions {
  databases: string[]
  schemas: string[]
  commandTypes: string[]
}

export interface DashboardSummary {
  totalCommands: number
  totalErrors: number
  avgDurationSeconds: number
  totalIndexRebuilds: number
  totalIndexReorganizes: number
  totalStatisticsUpdates: number
}

export interface StatisticsRow {
  database: string
  schema: string
  object: string
  updateCount: number
  lastUpdated: string
  avgDurationSeconds: number
  errorCount: number
  lastPageCount: number | null
  lastModificationCounter: number | null
  maxModificationCounter: number | null
}

export interface MaintenanceSummaryRow {
  commandType: string
  totalOperations: number
  successes: number
  errorCount: number
  totalDurationSeconds: number
  avgDurationSeconds: number
  firstOperation: string
  lastOperation: string
}

export interface TimelinePoint {
  date: string
  commandType: string
  totalDurationSeconds: number
}

export interface MaintenanceReport {
  summary: MaintenanceSummaryRow[]
  timeline: TimelinePoint[]
}

export interface ErrorSummaryRow {
  commandType: string
  errorCount: number
  firstError: string
  lastError: string
}

export interface ErrorTimelinePoint {
  date: string
  commandType: string
  errorCount: number
}

export interface ErrorDetailRow {
  database: string
  schema: string
  object: string
  index: string
  commandType: string
  startTime: string
  durationSeconds: number | null
  errorNumber: number
  errorMessage: string | null
}

export interface MaintenanceErrorsReport {
  summary: ErrorSummaryRow[]
  timeline: ErrorTimelinePoint[]
  detail: ErrorDetailRow[]
}

export interface LongRunningRow {
  database: string
  schema: string
  object: string
  index: string
  commandType: string
  startTime: string
  durationSeconds: number
  errorNumber: number | null
  errorMessage: string | null
}

export interface LongRunningTimelinePoint {
  date: string
  commandType: string
  maxDurationSeconds: number
  avgDurationSeconds: number
  count: number
}

export interface LongRunningReport {
  operations: LongRunningRow[]
  timeline: LongRunningTimelinePoint[]
}

export interface OperationsBatchPoint {
  date: string
  commandType: string
  count: number
  errorCount: number
}

export interface OperationsSummaryRow {
  commandType: string
  totalCount: number
  successCount: number
  errorCount: number
  firstOperation: string
  lastOperation: string
}

export interface OperationsBatchReport {
  summary: OperationsSummaryRow[]
  timeline: OperationsBatchPoint[]
}

export interface BackupTimelinePoint {
  date: string
  database: string
  totalCount: number
  successCount: number
  errorCount: number
  totalDurationSeconds: number
  avgDurationSeconds: number
}

export interface BackupReport {
  timeline: BackupTimelinePoint[]
}

export interface IndexRow {
  database: string
  schema: string
  object: string
  index: string
  rebuildCount: number
  reorganizeCount: number
  totalOperations: number
  lastOperation: string
  avgDurationSeconds: number
  errorCount: number
  maxFragmentation: number | null
  lastFragmentation: number | null
}
