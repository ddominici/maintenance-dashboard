import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import { getMaintenanceReport } from '../../api/maintenance'
import { getFilterOptions } from '../../api/dashboard'
import type { MaintenanceSummaryRow, TimelinePoint } from '../../types/api'

type Granularity = 'day' | 'week' | 'month'

// Fixed palette: cycles if there are more than 8 command types.
const LINE_COLORS = [
  '#f59e0b', '#3b82f6', '#10b981', '#8b5cf6',
  '#ef4444', '#06b6d4', '#84cc16', '#f97316',
]

function formatDuration(totalSeconds: number): string {
  if (totalSeconds < 60) return `${totalSeconds}s`
  const h = Math.floor(totalSeconds / 3600)
  const m = Math.floor((totalSeconds % 3600) / 60)
  const s = totalSeconds % 60
  if (h > 0) return `${h}h ${m}m ${s}s`
  return `${m}m ${s}s`
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    day: '2-digit', month: '2-digit', year: 'numeric',
  })
}

function Th({ children, right }: { children: React.ReactNode; right?: boolean }) {
  return (
    <th className={`whitespace-nowrap px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500 ${right ? 'text-right' : 'text-left'}`}>
      {children}
    </th>
  )
}

function SummaryRow({ row, index }: { row: MaintenanceSummaryRow; index: number }) {
  const base = index % 2 === 0 ? 'bg-white' : 'bg-slate-50'
  const errorRate = row.totalOperations > 0
    ? ((row.errorCount / row.totalOperations) * 100).toFixed(1)
    : '0.0'
  return (
    <tr className={`${base} hover:bg-blue-50 transition-colors`}>
      <td className="px-4 py-3 text-sm font-medium text-slate-900">{row.commandType || '—'}</td>
      <td className="px-4 py-3 text-right text-sm font-semibold text-slate-900">{row.totalOperations.toLocaleString()}</td>
      <td className="px-4 py-3 text-right text-sm text-emerald-600">{row.successes.toLocaleString()}</td>
      <td className="px-4 py-3 text-right text-sm">
        {row.errorCount > 0
          ? <span className="font-semibold text-red-600">{row.errorCount.toLocaleString()}</span>
          : <span className="text-slate-400">0</span>}
      </td>
      <td className="px-4 py-3 text-right text-sm text-slate-500">{errorRate}%</td>
      <td className="px-4 py-3 text-right text-sm text-slate-600">{formatDuration(row.totalDurationSeconds)}</td>
      <td className="px-4 py-3 text-right text-sm text-slate-600">{row.avgDurationSeconds.toFixed(1)}s</td>
      <td className="whitespace-nowrap px-4 py-3 text-sm text-slate-500">{formatDate(row.firstOperation)}</td>
      <td className="whitespace-nowrap px-4 py-3 text-sm text-slate-500">{formatDate(row.lastOperation)}</td>
    </tr>
  )
}

// ── Chart helpers ─────────────────────────────────────────────────────────────

function buildChartData(timeline: TimelinePoint[]) {
  const commandTypeSet = new Set<string>()
  const byDate = new Map<string, Record<string, unknown>>()

  for (const p of timeline) {
    commandTypeSet.add(p.commandType)
    if (!byDate.has(p.date)) byDate.set(p.date, { date: p.date })
    // Convert seconds → minutes for a more readable Y axis.
    byDate.get(p.date)![p.commandType] = p.totalDurationSeconds / 60
  }

  return {
    chartData: Array.from(byDate.values()).sort((a, b) =>
      String(a.date).localeCompare(String(b.date))
    ),
    commandTypes: Array.from(commandTypeSet).sort(),
  }
}

function formatXTick(date: string, granularity: Granularity): string {
  if (granularity === 'month') return date  // "2024-01"
  return date.slice(5)                      // "01-15" for day / week
}

function CustomTooltip({ active, payload, label }: {
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
          <span className="font-medium">{entry.value.toFixed(1)} min</span>
        </div>
      ))}
    </div>
  )
}

// ── Page ──────────────────────────────────────────────────────────────────────

export function MaintenanceSummaryPage() {
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
    queryKey: ['maintenance-summary', database, dateFrom, dateTo, granularity],
    queryFn: () => getMaintenanceReport(queryString),
  })

  const { chartData, commandTypes } = useMemo(
    () => buildChartData(reportQuery.data?.data.timeline ?? []),
    [reportQuery.data],
  )

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold">Maintenance Summary</h2>
        <p className="text-slate-600">Overview of all maintenance operations grouped by type.</p>
      </div>

      {/* Filters */}
      <div className="grid grid-cols-1 gap-4 rounded-lg bg-white p-4 shadow-sm ring-1 ring-slate-200 sm:grid-cols-2 xl:grid-cols-4">
        <label className="space-y-1">
          <span className="text-sm font-medium">Database</span>
          <select value={database} onChange={(e) => setDatabase(e.target.value)}
            className="w-full rounded border px-3 py-2 text-sm">
            <option value="">All</option>
            {filtersQuery.data?.data.databases.map((x) => (
              <option key={x} value={x}>{x}</option>
            ))}
          </select>
        </label>

        <label className="space-y-1">
          <span className="text-sm font-medium">Date from</span>
          <input type="date" value={dateFrom} onChange={(e) => setDateFrom(e.target.value)}
            className="w-full rounded border px-3 py-2 text-sm" />
        </label>

        <label className="space-y-1">
          <span className="text-sm font-medium">Date to</span>
          <input type="date" value={dateTo} onChange={(e) => setDateTo(e.target.value)}
            className="w-full rounded border px-3 py-2 text-sm" />
        </label>

        <label className="space-y-1">
          <span className="text-sm font-medium">Granularity</span>
          <select value={granularity} onChange={(e) => setGranularity(e.target.value as Granularity)}
            className="w-full rounded border px-3 py-2 text-sm">
            <option value="day">Daily</option>
            <option value="week">Weekly</option>
            <option value="month">Monthly</option>
          </select>
        </label>
      </div>

      {/* Loading */}
      {reportQuery.isLoading && (
        <div className="rounded-lg bg-white p-8 text-center text-slate-500 shadow-sm ring-1 ring-slate-200">
          Loading maintenance data…
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
          {/* ── Line chart ──────────────────────────────────────────────────── */}
          <div className="rounded-lg bg-white p-4 shadow-sm ring-1 ring-slate-200">
            <h3 className="mb-4 text-sm font-semibold text-slate-700">
              Time spent per maintenance type (minutes)
            </h3>
            {chartData.length === 0 ? (
              <div className="py-12 text-center text-sm text-slate-400">No data for the selected period.</div>
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={chartData} margin={{ top: 4, right: 16, left: 0, bottom: 4 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis
                    dataKey="date"
                    tick={{ fontSize: 11, fill: '#94a3b8' }}
                    tickFormatter={(d) => formatXTick(String(d), granularity)}
                    interval="preserveStartEnd"
                  />
                  <YAxis
                    tick={{ fontSize: 11, fill: '#94a3b8' }}
                    tickFormatter={(v: number) => `${v.toFixed(0)}m`}
                    width={48}
                  />
                  <Tooltip content={<CustomTooltip />} />
                  <Legend
                    wrapperStyle={{ fontSize: 12, paddingTop: 12 }}
                    formatter={(value) => <span className="text-slate-600">{value}</span>}
                  />
                  {commandTypes.map((type, i) => (
                    <Line
                      key={type}
                      type="monotone"
                      dataKey={type}
                      stroke={LINE_COLORS[i % LINE_COLORS.length]}
                      strokeWidth={2}
                      dot={false}
                      activeDot={{ r: 4 }}
                    />
                  ))}
                </LineChart>
              </ResponsiveContainer>
            )}
          </div>

          {/* ── Summary table ────────────────────────────────────────────────── */}
          <div className="overflow-hidden rounded-lg bg-white shadow-sm ring-1 ring-slate-200">
            {reportQuery.data.data.summary.length === 0 ? (
              <div className="p-8 text-center text-slate-500">No maintenance operations found.</div>
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-slate-200">
                  <thead className="bg-slate-50">
                    <tr>
                      <Th>Command Type</Th>
                      <Th right>Total</Th>
                      <Th right>OK</Th>
                      <Th right>Errors</Th>
                      <Th right>Error %</Th>
                      <Th right>Total Duration</Th>
                      <Th right>Avg Duration</Th>
                      <Th>First Op</Th>
                      <Th>Last Op</Th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {reportQuery.data.data.summary.map((row, i) => (
                      <SummaryRow key={row.commandType} row={row} index={i} />
                    ))}
                  </tbody>
                </table>
              </div>
            )}
            <div className="border-t border-slate-100 px-4 py-2 text-right text-xs text-slate-400">
              {reportQuery.data.data.summary.length} command types
            </div>
          </div>
        </>
      )}
    </div>
  )
}
