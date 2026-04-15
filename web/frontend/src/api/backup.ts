import { getJson } from './client'
import type { ApiEnvelope, BackupReport } from '../types/api'

export function getBackupReport(query: string) {
  return getJson<ApiEnvelope<BackupReport>>(`/api/backup/report${query}`)
}
