import { getJson } from './client'
import type { ApiEnvelope, DashboardSummary, FilterOptions } from '../types/api'

export function getDashboardSummary(query: string) {
  return getJson<ApiEnvelope<DashboardSummary>>(`/api/dashboard/summary${query}`)
}

export function getFilterOptions() {
  return getJson<ApiEnvelope<FilterOptions>>('/api/meta/filters')
}
