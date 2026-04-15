import { getJson } from './client'
import type { ApiEnvelope, MaintenanceReport } from '../types/api'

export function getMaintenanceReport(query: string) {
  return getJson<ApiEnvelope<MaintenanceReport>>(`/api/maintenance/summary${query}`)
}
