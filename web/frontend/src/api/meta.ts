import { getJson } from './client'
import type { ApiEnvelope } from '../types/api'

export interface ServerStatus {
  name: string
  host: string
  reachable: boolean
}

export function getServers() {
  return getJson<ApiEnvelope<string[]>>('/api/meta/servers')
}

export async function getServerStatus(): Promise<ApiEnvelope<ServerStatus[]>> {
  const res = await fetch('/api/meta/server-status', { headers: { Accept: 'application/json' } })
  if (!res.ok) throw new Error(`Request failed with status ${res.status}`)
  return res.json() as Promise<ApiEnvelope<ServerStatus[]>>
}
