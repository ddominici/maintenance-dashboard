import { getJson } from './client'
import type { ApiEnvelope, LongRunningReport } from '../types/api'

export function getLongRunningReport(query: string) {
  return getJson<ApiEnvelope<LongRunningReport>>(`/api/longrunning/report${query}`)
}
