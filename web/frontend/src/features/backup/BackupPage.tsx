import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import { getBackupReport } from '../../api/backup'
import { getFilterOptions } from '../../api/dashboard'
import type { BackupTimelinePoint } from '../../types/api'

type Granularity = 'day' | 'week' | 'month'

const BAR_COLORS = [
  '#3b82f6', '#10b981', '#f59e0b', '#8b5cf6',
  '#ef4444', '#06b6d4', '#84cc16', '#f97316',
]

// ── Data helpers ──────────────────────────────────────────────────────────────

interface PivotResult {
  chartData: Record<string, unknown>[]
  databases: string[]
}

function pivotTimeline(
  timeline: BackupTimelinePoint[],
  valueKey: keyof BackupTimelinePoint,
): PivotResult {
  const dbSet = new Set<string>()
  const byDate = new Map<string, Record<string, unknown>>()

  for (const p of timeline) {
    dbSet.add(p.database)
    if (!byDate.has(p.date)) byDate.set(p.date, { date: p.date })
    byDate.get(p.date)![p.database] = p[valueKey]
  }

  return {
    chartData: Array.from(byDate.values()).sort((a, b) =>
      String(a.date).localeCompare(String(b.date)),
    ),
    databases: Array.from(dbSet).sort(),
  }
}

function formatXTick(date: string, granularity: Granularity): string {
  if (granularity === 'month') return date
  return date.slice(5)
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  if (h > 0) return `${h}h ${m}m ${s}s`
  return `${m}m ${s}s`
}

// ── Custom tooltips ───────────────────────────────────────────────────────────

function DurationTooltip({ active, payload, label }: {
  active?: boolean
  payload?: Array<{ name: string; value: number; color: string }>
  label?: string
}) {
  if (!active || !payload?.length) return null
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-3 shadow-lg text-xs">
      <p className="mb-2 font-semibold text-slate-700">{label}</p>
      {payload.map((entry) => (
        <div key={entry.name} className="flex items-center gap-2">
          <span className="inline-block h-2 w-2 rounded-full" style={{ background: entry.color }} />
          <span className="text-slate-600">{entry.name}:</span>
          <span className="font-medium">{formatDuration(entry.value)}</span>
        </div>
      ))}
    </div>
  )
}

function ErrorTooltip({ active, payload, label }: {
  active?: boolean
  payload?: Array<{ name: string; value: number; color: string }>
  label?: string
}) {
  if (!active || !payload?.length) return null
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-3 shadow-lg text-xs">
      <p className="mb-2 font-semibold text-slate-700">{label}</p>
      {payload.map((entry) => (
        <div key={entry.name} className="flex items-center gap-2">
          <span className="inline-block h-2 w-2 rounded-full" style={{ background: entry.color }} />
          <span className="text-slate-600">{entry.name}:</span>
          <span className="font-medium text-red-600">{entry.value} error{entry.value !== 1 ? 's' : ''}</span>
        </div>
      ))}
    </div>
  )
}

// ── Summary table ─────────────────────────────────────────────────────────────

function buildSummary(timeline: BackupTimelinePoint[]) {
  const map = new Map<string, {
    totalCount: number
    successCount: number
    errorCount: number
    totalDurationSeconds: number
  }>()

  for (const p of timeline) {
    const existing = map.get(p.database) ?? {
      totalCount: 0, successCount: 0, errorCount: 0, totalDurationSeconds: 0,
    }
    map.set(p.database, {
      totalCount: existing.totalCount + p.totalCount,
      successCount: existing.successCount + p.successCount,
      errorCount: existing.errorCount + p.errorCount,
      totalDurationSeconds: existing.totalDurationSeconds + p.totalDurationSeconds,
    })
  }

  return Array.from(map.entries())
    .map(([database, v]) => ({ database, ...v }))
    .sort((a, b) => a.database.localeCompare(b.database))
}

function Th({ children, right }: { children: React.ReactNode; right?: boolean }) {
  return (
    <th className={`whitespace-nowrap px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500 ${right ? 'text-right' : 'text-left'}`}>
      {children}
    </th>
  )
}

// ── Page ──────────────────────────────────────────────────────────────────────

export function BackupPage() {
  const [database, setDatabase] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [granularity, setGranularity] = useState<Granularity>('day')

  const queryString = useMemo(() => {
    const p = new URLSearchParams()
    if (database) p.set('database', database)
    if (dateFrom) p.set('dateFrom', dateFrom)
    if (dateTo) p.set('dateTo', dateTo)
    p.set('granularity', granularity)
    return `?${p.toString()}`
  }, [database, dateFrom, dateTo, granularity])

  const filtersQuery = useQuery({ queryKey: ['filters'], queryFn: getFilterOptions })
  const reportQuery = useQuery({
    queryKey: ['backup-report', database, dateFrom, dateTo, granularity],
    queryFn: () => getBackupReport(queryString),
  })

  const timeline = reportQuery.data?.data.timeline ?? []

  const { chartData: durationData, databases } = useMemo(
    () => pivotTimeline(timeline, 'totalDurationSeconds'),
    [timeline],
  )
  const { chartData: errorData } = useMemo(
    () => pivotTimeline(timeline, 'errorCount'),
    [timeline],
  )
  const summary = useMemo(() => buildSummary(timeline), [timeline])

  const hasErrors = summary.some((r) => r.errorCount > 0)

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold">Backup Overview</h2>
        <p className="text-slate-600">Duration and error trends for backup operations.</p>
      </div>

      {/* Filters */}
      <div className="grid grid-cols-1 gap-4 rounded-lg bg-white p-4 shadow-sm ring-1 ring-slate-200 sm:grid-cols-2 xl:grid-cols-4">
        <label className="space-y-1">
          <span className="text-sm font-medium">Database</span>
          <select
            value={database}
            onChange={(e) => setDatabase(e.target.value)}
            className="w-full rounded border px-3 py-2 text-sm"
          >
            <option value="">All</option>
            {filtersQuery.data?.data.databases.map((x) => (
              <option key={x} value={x}>{x}</option>
            ))}
          </select>
        </label>

        <label className="space-y-1">
          <span className="text-sm font-medium">Date from</span>
          <input
            type="date"
            value={dateFrom}
            onChange={(e) => setDateFrom(e.target.value)}
            className="w-full rounded border px-3 py-2 text-sm"
          />
        </label>

        <label className="space-y-1">
          <span className="text-sm font-medium">Date to</span>
          <input
            type="date"
            value={dateTo}
            onChange={(e) => setDateTo(e.target.value)}
            className="w-full rounded border px-3 py-2 text-sm"
          />
        </label>

        <label className="space-y-1">
          <span className="text-sm font-medium">Granularity</span>
          <select
            value={granularity}
            onChange={(e) => setGranularity(e.target.value as Granularity)}
            className="w-full rounded border px-3 py-2 text-sm"
          >
            <option value="day">Daily</option>
            <option value="week">Weekly</option>
            <option value="month">Monthly</option>
          </select>
        </label>
      </div>

      {/* Loading */}
      {reportQuery.isLoading && (
        <div className="rounded-lg bg-white p-8 text-center text-slate-500 shadow-sm ring-1 ring-slate-200">
          Loading backup data…
        </div>
      )}

      {/* Error */}
      {reportQuery.isError && (
        <div className="rounded border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {(reportQuery.error as Error).message}
        </div>
      )}

      {reportQuery.data && (
        <>
          {/* ── Duration chart ──────────────────────────────────────────────── */}
          <div className="rounded-lg bg-white p-4 shadow-sm ring-1 ring-slate-200">
            <h3 className="mb-4 text-sm font-semibold text-slate-700">
              Backup duration (seconds)
            </h3>
            {durationData.length === 0 ? (
              <div className="py-12 text-center text-sm text-slate-400">
                No backup operations found for the selected period.
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={durationData} margin={{ top: 4, right: 16, left: 0, bottom: 4 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis
                    dataKey="date"
                    tick={{ fontSize: 11, fill: '#94a3b8' }}
                    tickFormatter={(d) => formatXTick(String(d), granularity)}
                    interval="preserveStartEnd"
                  />
                  <YAxis
                    tick={{ fontSize: 11, fill: '#94a3b8' }}
                    tickFormatter={(v: number) => `${v}s`}
                    width={56}
                  />
                  <Tooltip content={<DurationTooltip />} />
                  <Legend
                    wrapperStyle={{ fontSize: 12, paddingTop: 12 }}
                    formatter={(value) => <span className="text-slate-600">{value}</span>}
                  />
                  {databases.map((db, i) => (
                    <Bar
                      key={db}
                      dataKey={db}
                      stackId="duration"
                      fill={BAR_COLORS[i % BAR_COLORS.length]}
                      maxBarSize={40}
                    />
                  ))}
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>

          {/* ── Error chart ─────────────────────────────────────────────────── */}
          <div className="rounded-lg bg-white p-4 shadow-sm ring-1 ring-slate-200">
            <h3 className="mb-1 text-sm font-semibold text-slate-700">
              Failed backups
            </h3>
            {!hasErrors ? (
              <div className="py-12 text-center text-sm text-emerald-600 font-medium">
                No backup errors in the selected period.
              </div>
            ) : (
              <ResponsiveContainer width="100%" height={240}>
                <BarChart data={errorData} margin={{ top: 4, right: 16, left: 0, bottom: 4 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis
                    dataKey="date"
                    tick={{ fontSize: 11, fill: '#94a3b8' }}
                    tickFormatter={(d) => formatXTick(String(d), granularity)}
                    interval="preserveStartEnd"
                  />
                  <YAxis
                    allowDecimals={false}
                    tick={{ fontSize: 11, fill: '#94a3b8' }}
                    width={36}
                  />
                  <Tooltip content={<ErrorTooltip />} />
                  <Legend
                    wrapperStyle={{ fontSize: 12, paddingTop: 12 }}
                    formatter={(value) => <span className="text-slate-600">{value}</span>}
                  />
                  {databases.map((db, i) => (
                    <Bar
                      key={db}
                      dataKey={db}
                      stackId="errors"
                      fill={BAR_COLORS[i % BAR_COLORS.length]}
                      maxBarSize={40}
                    />
                  ))}
                </BarChart>
              </ResponsiveContainer>
            )}
          </div>

          {/* ── Summary table ────────────────────────────────────────────────── */}
          <div className="overflow-hidden rounded-lg bg-white shadow-sm ring-1 ring-slate-200">
            {summary.length === 0 ? (
              <div className="p-8 text-center text-slate-500">No backup operations found.</div>
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-slate-200">
                  <thead className="bg-slate-50">
                    <tr>
                      <Th>Database</Th>
                      <Th right>Total</Th>
                      <Th right>OK</Th>
                      <Th right>Errors</Th>
                      <Th right>Error %</Th>
                      <Th right>Total Duration</Th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {summary.map((row, i) => {
                      const errorRate = row.totalCount > 0
                        ? ((row.errorCount / row.totalCount) * 100).toFixed(1)
                        : '0.0'
                      const base = i % 2 === 0 ? 'bg-white' : 'bg-slate-50'
                      return (
                        <tr key={row.database} className={`${base} hover:bg-blue-50 transition-colors`}>
                          <td className="px-4 py-3 text-sm font-medium text-slate-900">{row.database || '—'}</td>
                          <td className="px-4 py-3 text-right text-sm font-semibold text-slate-900">{row.totalCount.toLocaleString()}</td>
                          <td className="px-4 py-3 text-right text-sm text-emerald-600">{row.successCount.toLocaleString()}</td>
                          <td className="px-4 py-3 text-right text-sm">
                            {row.errorCount > 0
                              ? <span className="font-semibold text-red-600">{row.errorCount.toLocaleString()}</span>
                              : <span className="text-slate-400">0</span>}
                          </td>
                          <td className="px-4 py-3 text-right text-sm text-slate-500">{errorRate}%</td>
                          <td className="px-4 py-3 text-right text-sm text-slate-600">{formatDuration(row.totalDurationSeconds)}</td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            )}
            <div className="border-t border-slate-100 px-4 py-2 text-right text-xs text-slate-400">
              {summary.length} database{summary.length !== 1 ? 's' : ''}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
