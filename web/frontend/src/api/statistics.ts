import { getJson } from './client'
import type { ApiEnvelope, StatisticsRow } from '../types/api'

export function getMostModifiedStatistics(query: string) {
  return getJson<ApiEnvelope<StatisticsRow[]>>(`/api/statistics/most-modified${query}`)
}
