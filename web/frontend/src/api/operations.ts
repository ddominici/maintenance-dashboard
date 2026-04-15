import { getJson } from './client'
import type { ApiEnvelope, OperationsBatchReport } from '../types/api'

export function getOperationsBatchReport(query: string) {
  return getJson<ApiEnvelope<OperationsBatchReport>>(`/api/operations/per-batch${query}`)
}
