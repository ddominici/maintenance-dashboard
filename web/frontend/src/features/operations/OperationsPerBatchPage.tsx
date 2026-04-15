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
import { getOperationsBatchReport } from '../../api/operations'
import { getFilterOptions } from '../../api/dashboard'
import type { OperationsBatchPoint, OperationsSummaryRow } from '../../types/api'

type Granularity = 'day' | 'week' | 'month'

const BAR_COLORS = [
  '#f59e0b', '#3b82f6', '#10b981', '#8b5cf6',
  '#ef4444', '#06b6d4', '#84cc16', '#f97316',
]

// ── Chart helpers ─────────────────────────────────────────────────────────────

function buildChartData(timeline: OperationsBatchPoint[]) {
  const commandTypeSet = new Set<string>()
  const byDate = new Map<string, Record<string, unknown>>()

  for (const p of timeline) {
    commandTypeSet.add(p.commandType)
    if (!byDate.has(p.date)) byDate.set(p.date, { date: p.date })
    byDate.get(p.date)![p.commandType] = p.count
  }

  return {
    chartData: Array.from(byDate.values()).sort((a, b) =>
      String(a.date).localeCompare(String(b.date)),
    ),
    commandTypes: Array.from(commandTypeSet).sort(),
  }
}

// Per-period stats used to compute avg and peak in the summary table.
function periodStats(timeline: OperationsBatchPoint[]): Map<string, { total: number; peak: number }> {
  // group total count per (commandType, date), then derive avg and peak per commandType
  const perTypeDate = new Map<string, Map<string, number>>()
  for (const p of timeline) {
    if (!perTypeDate.has(p.commandType)) perTypeDate.set(p.commandType, new Map())
    const dateMap = perTypeDate.get(p.commandType)!
    dateMap.set(p.date, (dateMap.get(p.date) ?? 0) + p.count)
  }
  const result = new Map<string, { total: number; peak: number }>()
  for (const [ct, dateMap] of perTypeDate) {
    const values = Array.from(dateMap.values())
    const peak = Math.max(...values)
    const avg = values.reduce((a, b) => a + b, 0) / values.length
    result.set(ct, { total: avg, peak })
  }
  return result
}

function formatXTick(date: string, granularity: Granularity): string {
  if (granularity === 'month') return date
  return date.slice(5)
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    day: '2-digit', month: '2-digit', year: 'numeric',
  })
}

function CustomTooltip({ active, payload, label }: {
  active?: boolean
  payload?: Array<{ name: string; value: number; color: string }>
  label?: string
}) {
  if (!active || !payload?.length) return null
  const total = payload.reduce((s, e) => s + (e.value ?? 0), 0)
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-3 shadow-lg text-xs">
      <p className="mb-2 font-semibold text-slate-700">{label}</p>
      {payload.map((entry) => (
        <div key={entry.name} className="flex items-center gap-2">
          <span className="inline-block h-2 w-2 rounded-full" style={{ background: entry.color }} />
          <span className="text-slate-600">{entry.name}:</span>
          <span className="font-medium">{entry.value.toLocaleString()}</span>
        </div>
      ))}
      {payload.length > 1 && (
        <div className="mt-1 border-t border-slate-100 pt-1 font-semibold text-slate-700">
          Total: {total.toLocaleString()}
        </div>
      )}
    </div>
  )
}

// ── Table ─────────────────────────────────────────────────────────────────────

function Th({ children, right }: { children: React.ReactNode; right?: boolean }) {
  return (
    <th className={`whitespace-nowrap px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500 ${right ? 'text-right' : 'text-left'}`}>
      {children}
    </th>
  )
}

function SummaryRow({
  row,
  index,
  stats,
}: {
  row: OperationsSummaryRow
  index: number
  stats: { total: number; peak: number } | undefined
}) {
  const base = index % 2 === 0 ? 'bg-white' : 'bg-slate-50'
  const errorRate = row.totalCount > 0
    ? ((row.errorCount / row.totalCount) * 100).toFixed(1)
    : '0.0'
  return (
    <tr className={`${base} hover:bg-blue-50 transition-colors`}>
      <td className="px-4 py-3 text-sm font-medium text-slate-900">{row.commandType || '—'}</td>
      <td className="px-4 py-3 text-right text-sm font-semibold text-slate-900">{row.totalCount.toLocaleString()}</td>
      <td className="px-4 py-3 text-right text-sm text-emerald-600">{row.successCount.toLocaleString()}</td>
      <td className="px-4 py-3 text-right text-sm">
        {row.errorCount > 0
          ? <span className="font-semibold text-red-600">{row.errorCount.toLocaleString()}</span>
          : <span className="text-slate-400">0</span>}
      </td>
      <td className="px-4 py-3 text-right text-sm text-slate-500">{errorRate}%</td>
      <td className="px-4 py-3 text-right text-sm text-slate-600">
        {stats ? stats.total.toFixed(1) : '—'}
      </td>
      <td className="px-4 py-3 text-right text-sm text-slate-600">
        {stats ? stats.peak.toLocaleString() : '—'}
      </td>
      <td className="whitespace-nowrap px-4 py-3 text-sm text-slate-500">{formatDate(row.firstOperation)}</td>
      <td className="whitespace-nowrap px-4 py-3 text-sm text-slate-500">{formatDate(row.lastOperation)}</td>
    </tr>
  )
}

// ── Page ──────────────────────────────────────────────────────────────────────

export function OperationsPerBatchPage() {
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
    queryKey: ['operations-per-batch', database, dateFrom, dateTo, granularity],
    queryFn: () => getOperationsBatchReport(queryString),
  })

  const { chartData, commandTypes } = useMemo(
    () => buildChartData(reportQuery.data?.data.timeline ?? []),
    [reportQuery.data],
  )

  const stats = useMemo(
    () => periodStats(reportQuery.data?.data.timeline ?? []),
    [reportQuery.data],
  )

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold">Operations Per Batch</h2>
        <p className="text-slate-600">Number of maintenance operations executed per period, grouped by command type.</p>
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
          Loading operations data…
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
          {/* ── Bar chart ───────────────────────────────────────────────────── */}
          <div className="rounded-lg bg-white p-4 shadow-sm ring-1 ring-slate-200">
            <h3 className="mb-4 text-sm font-semibold text-slate-700">
              Operations per period
            </h3>
            {chartData.length === 0 ? (
              <div className="py-12 text-center text-sm text-slate-400">No data for the selected period.</div>
            ) : (
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={chartData} margin={{ top: 4, right: 16, left: 0, bottom: 4 }}>
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
                    width={48}
                  />
                  <Tooltip content={<CustomTooltip />} />
                  <Legend
                    wrapperStyle={{ fontSize: 12, paddingTop: 12 }}
                    formatter={(value) => <span className="text-slate-600">{value}</span>}
                  />
                  {commandTypes.map((type, i) => (
                    <Bar
                      key={type}
                      dataKey={type}
                      stackId="ops"
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
            {reportQuery.data.data.summary.length === 0 ? (
              <div className="p-8 text-center text-slate-500">No operations found.</div>
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
                      <Th right>Avg / Period</Th>
                      <Th right>Peak / Period</Th>
                      <Th>First Op</Th>
                      <Th>Last Op</Th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {reportQuery.data.data.summary.map((row, i) => (
                      <SummaryRow
                        key={row.commandType}
                        row={row}
                        index={i}
                        stats={stats.get(row.commandType)}
                      />
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
