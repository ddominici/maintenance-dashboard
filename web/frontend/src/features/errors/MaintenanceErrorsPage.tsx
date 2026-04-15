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
import { getMaintenanceErrorsReport } from '../../api/errors'
import { getFilterOptions } from '../../api/dashboard'
import type { ErrorDetailRow, ErrorSummaryRow, ErrorTimelinePoint } from '../../types/api'

type Granularity = 'day' | 'week' | 'month'

const BAR_COLORS = [
  '#ef4444', '#f97316', '#f59e0b', '#8b5cf6',
  '#3b82f6', '#10b981', '#06b6d4', '#84cc16',
]

const LIMIT_OPTIONS = [25, 50, 100, 200]

// ── Helpers ───────────────────────────────────────────────────────────────────

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  if (h > 0) return `${h}h ${m}m ${s}s`
  return `${m}m ${s}s`
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    day: '2-digit', month: '2-digit', year: 'numeric',
  })
}

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString(undefined, {
    day: '2-digit', month: '2-digit', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
}

function formatXTick(date: string, granularity: Granularity): string {
  if (granularity === 'month') return date
  return date.slice(5)
}

// ── Chart helpers ─────────────────────────────────────────────────────────────

function buildChartData(timeline: ErrorTimelinePoint[]) {
  const commandTypeSet = new Set<string>()
  const byDate = new Map<string, Record<string, unknown>>()

  for (const p of timeline) {
    commandTypeSet.add(p.commandType)
    if (!byDate.has(p.date)) byDate.set(p.date, { date: p.date })
    byDate.get(p.date)![p.commandType] = p.errorCount
  }

  return {
    chartData: Array.from(byDate.values()).sort((a, b) =>
      String(a.date).localeCompare(String(b.date)),
    ),
    commandTypes: Array.from(commandTypeSet).sort(),
  }
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
          <span className="font-medium text-red-600">{entry.value}</span>
        </div>
      ))}
      {payload.length > 1 && (
        <div className="mt-1 border-t border-slate-100 pt-1 font-semibold text-red-600">
          Total: {total}
        </div>
      )}
    </div>
  )
}

// ── Tables ────────────────────────────────────────────────────────────────────

function Th({ children, right }: { children: React.ReactNode; right?: boolean }) {
  return (
    <th className={`whitespace-nowrap px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500 ${right ? 'text-right' : 'text-left'}`}>
      {children}
    </th>
  )
}

function SummaryRow({ row, index }: { row: ErrorSummaryRow; index: number }) {
  const base = index % 2 === 0 ? 'bg-white' : 'bg-slate-50'
  return (
    <tr className={`${base} hover:bg-red-50 transition-colors`}>
      <td className="px-4 py-3 text-sm font-medium text-slate-900">{row.commandType || '—'}</td>
      <td className="px-4 py-3 text-right text-sm font-semibold text-red-600">{row.errorCount.toLocaleString()}</td>
      <td className="whitespace-nowrap px-4 py-3 text-sm text-slate-500">{formatDate(row.firstError)}</td>
      <td className="whitespace-nowrap px-4 py-3 text-sm text-slate-500">{formatDate(row.lastError)}</td>
    </tr>
  )
}

function DetailRow({ row, index }: { row: ErrorDetailRow; index: number }) {
  const base = index % 2 === 0 ? 'bg-white' : 'bg-slate-50'
  // Truncate long error messages; full text via title tooltip.
  const msgShort = row.errorMessage && row.errorMessage.length > 80
    ? row.errorMessage.slice(0, 80) + '…'
    : row.errorMessage

  return (
    <tr className={`${base} hover:bg-red-50 transition-colors`}>
      <td className="px-4 py-3 text-sm text-slate-500">{row.database || '—'}</td>
      <td className="px-4 py-3 text-sm text-slate-500">{row.schema || '—'}</td>
      <td className="px-4 py-3 text-sm font-medium text-slate-900">{row.object || '—'}</td>
      <td className="px-4 py-3 text-sm text-slate-600">{row.index || '—'}</td>
      <td className="px-4 py-3 text-sm text-slate-700">{row.commandType || '—'}</td>
      <td className="whitespace-nowrap px-4 py-3 text-sm text-slate-600">{formatDateTime(row.startTime)}</td>
      <td className="px-4 py-3 text-right text-sm text-slate-500">
        {row.durationSeconds !== null ? formatDuration(row.durationSeconds) : <span className="text-slate-300">—</span>}
      </td>
      <td className="px-4 py-3 text-right text-sm font-semibold text-red-600">
        #{row.errorNumber}
      </td>
      <td
        className="max-w-xs px-4 py-3 text-sm text-slate-600"
        title={row.errorMessage ?? undefined}
      >
        {msgShort ?? <span className="text-slate-300">—</span>}
      </td>
    </tr>
  )
}

// ── Page ──────────────────────────────────────────────────────────────────────

export function MaintenanceErrorsPage() {
  const [database, setDatabase]       = useState('')
  const [dateFrom, setDateFrom]       = useState('')
  const [dateTo, setDateTo]           = useState('')
  const [commandType, setCommandType] = useState('')
  const [granularity, setGranularity] = useState<Granularity>('day')
  const [limit, setLimit]             = useState(100)

  const queryString = useMemo(() => {
    const p = new URLSearchParams()
    if (database)    p.set('database', database)
    if (dateFrom)    p.set('dateFrom', dateFrom)
    if (dateTo)      p.set('dateTo', dateTo)
    if (commandType) p.set('commandType', commandType)
    p.set('granularity', granularity)
    p.set('limit', String(limit))
    return `?${p.toString()}`
  }, [database, dateFrom, dateTo, commandType, granularity, limit])

  const filtersQuery = useQuery({ queryKey: ['filters'], queryFn: getFilterOptions })
  const reportQuery  = useQuery({
    queryKey: ['maintenance-errors', database, dateFrom, dateTo, commandType, granularity, limit],
    queryFn: () => getMaintenanceErrorsReport(queryString),
  })

  const { chartData, commandTypes } = useMemo(
    () => buildChartData(reportQuery.data?.data.timeline ?? []),
    [reportQuery.data],
  )

  const totalErrors = reportQuery.data?.data.summary.reduce((s, r) => s + r.errorCount, 0) ?? 0

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold">Maintenance Errors</h2>
        <p className="text-slate-600">
          All maintenance operations that completed with an error, grouped by command type.
        </p>
      </div>

      {/* Filters */}
      <div className="grid grid-cols-1 gap-4 rounded-lg bg-white p-4 shadow-sm ring-1 ring-slate-200 sm:grid-cols-2 xl:grid-cols-3">
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
          <span className="text-sm font-medium">Command type</span>
          <select value={commandType} onChange={(e) => setCommandType(e.target.value)}
            className="w-full rounded border px-3 py-2 text-sm">
            <option value="">All</option>
            {filtersQuery.data?.data.commandTypes.map((x) => (
              <option key={x} value={x}>{x}</option>
            ))}
          </select>
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
      </div>

      {/* Loading */}
      {reportQuery.isLoading && (
        <div className="rounded-lg bg-white p-8 text-center text-slate-500 shadow-sm ring-1 ring-slate-200">
          Loading error data…
        </div>
      )}

      {/* Fetch error */}
      {reportQuery.isError && (
        <div className="rounded border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {(reportQuery.error as Error).message}
        </div>
      )}

      {reportQuery.data && totalErrors === 0 && (
        <div className="rounded-lg bg-white p-12 text-center shadow-sm ring-1 ring-slate-200">
          <p className="text-lg font-medium text-emerald-600">No errors found</p>
          <p className="mt-1 text-sm text-slate-500">All maintenance operations completed successfully for the selected period.</p>
        </div>
      )}

      {reportQuery.data && totalErrors > 0 && (
        <>
          {/* ── Bar chart ───────────────────────────────────────────────────── */}
          <div className="rounded-lg bg-white p-4 shadow-sm ring-1 ring-slate-200">
            <h3 className="mb-4 text-sm font-semibold text-slate-700">
              Errors per period
            </h3>
            {chartData.length === 0 ? (
              <div className="py-12 text-center text-sm text-slate-400">No data for the selected period.</div>
            ) : (
              <ResponsiveContainer width="100%" height={280}>
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
                    width={36}
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
            <div className="border-b border-slate-100 px-4 py-3">
              <h3 className="text-sm font-semibold text-slate-700">Errors by command type</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-slate-200">
                <thead className="bg-slate-50">
                  <tr>
                    <Th>Command Type</Th>
                    <Th right>Error Count</Th>
                    <Th>First Error</Th>
                    <Th>Last Error</Th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {reportQuery.data.data.summary.map((row, i) => (
                    <SummaryRow key={row.commandType} row={row} index={i} />
                  ))}
                </tbody>
              </table>
            </div>
            <div className="border-t border-slate-100 px-4 py-2 text-right text-xs text-slate-400">
              {totalErrors.toLocaleString()} total errors · {reportQuery.data.data.summary.length} command types
            </div>
          </div>

          {/* ── Detail table ─────────────────────────────────────────────────── */}
          <div className="overflow-hidden rounded-lg bg-white shadow-sm ring-1 ring-slate-200">
            <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
              <h3 className="text-sm font-semibold text-slate-700">Error detail</h3>
              <label className="flex items-center gap-2 text-sm text-slate-600">
                Show
                <select
                  value={limit}
                  onChange={(e) => setLimit(Number(e.target.value))}
                  className="rounded border px-2 py-1 text-sm"
                >
                  {LIMIT_OPTIONS.map((n) => (
                    <option key={n} value={n}>{n}</option>
                  ))}
                </select>
                most recent
              </label>
            </div>
            {reportQuery.data.data.detail.length === 0 ? (
              <div className="p-8 text-center text-slate-500">No detail rows available.</div>
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-slate-200">
                  <thead className="bg-slate-50">
                    <tr>
                      <Th>Database</Th>
                      <Th>Schema</Th>
                      <Th>Object</Th>
                      <Th>Index</Th>
                      <Th>Command Type</Th>
                      <Th>Start Time</Th>
                      <Th right>Duration</Th>
                      <Th right>Error #</Th>
                      <Th>Error Message</Th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {reportQuery.data.data.detail.map((row, i) => (
                      <DetailRow
                        key={`${row.database}-${row.schema}-${row.object}-${row.startTime}`}
                        row={row}
                        index={i}
                      />
                    ))}
                  </tbody>
                </table>
              </div>
            )}
            <div className="border-t border-slate-100 px-4 py-2 text-right text-xs text-slate-400">
              {reportQuery.data.data.detail.length} rows (most recent first)
            </div>
          </div>
        </>
      )}
    </div>
  )
}
