import { getJson } from './client'
import type { ApiEnvelope, IndexRow } from '../types/api'

export function getTopFragmentedIndexes(query: string) {
  return getJson<ApiEnvelope<IndexRow[]>>(`/api/indexes/top-fragmented${query}`)
}
