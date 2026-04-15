import { getJson } from './client'
import type { ApiEnvelope, MaintenanceErrorsReport } from '../types/api'

export function getMaintenanceErrorsReport(query: string) {
  return getJson<ApiEnvelope<MaintenanceErrorsReport>>(`/api/errors/report${query}`)
}
