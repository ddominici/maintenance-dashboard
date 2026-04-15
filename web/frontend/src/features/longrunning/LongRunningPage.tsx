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
import { getLongRunningReport } from '../../api/longrunning'
import { getFilterOptions } from '../../api/dashboard'
import type { LongRunningRow, LongRunningTimelinePoint } from '../../types/api'

type Granularity = 'day' | 'week' | 'month'

const LINE_COLORS = [
  '#f59e0b', '#3b82f6', '#10b981', '#8b5cf6',
  '#ef4444', '#06b6d4', '#84cc16', '#f97316',
]

const MIN_DURATION_OPTIONS = [
  { label: '1 min',   value: 60 },
  { label: '5 min',   value: 300 },
  { label: '15 min',  value: 900 },
  { label: '30 min',  value: 1800 },
  { label: '1 hour',  value: 3600 },
  { label: '2 hours', value: 7200 },
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

// ── Chart ─────────────────────────────────────────────────────────────────────

function buildChartData(timeline: LongRunningTimelinePoint[]) {
  const commandTypeSet = new Set<string>()
  const byDate = new Map<string, Record<string, unknown>>()

  for (const p of timeline) {
    commandTypeSet.add(p.commandType)
    if (!byDate.has(p.date)) byDate.set(p.date, { date: p.date })
    // Convert to minutes for a readable Y axis.
    byDate.get(p.date)![p.commandType] = p.maxDurationSeconds / 60
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
  return (
    <div className="rounded-lg border border-slate-200 bg-white p-3 shadow-lg text-xs">
      <p className="mb-2 font-semibold text-slate-700">{label}</p>
      {payload.map((entry) => (
        <div key={entry.name} className="flex items-center gap-2">
          <span className="inline-block h-2 w-2 rounded-full" style={{ background: entry.color }} />
          <span className="text-slate-600">{entry.name}:</span>
          <span className="font-medium">{formatDuration(Math.round(entry.value * 60))}</span>
        </div>
      ))}
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

function OpRow({ row, index }: { row: LongRunningRow; index: number }) {
  const base = index % 2 === 0 ? 'bg-white' : 'bg-slate-50'
  const hasError = row.errorNumber !== null && row.errorNumber !== 0

  // Duration bar: capped at 2 hours for visual proportion.
  const MAX_BAR_SECONDS = 7200
  const pct = Math.min((row.durationSeconds / MAX_BAR_SECONDS) * 100, 100)
  const barColor = row.durationSeconds >= 3600
    ? 'bg-red-500'
    : row.durationSeconds >= 900
      ? 'bg-amber-400'
      : 'bg-blue-400'

  return (
    <tr className={`${base} hover:bg-blue-50 transition-colors`}>
      <td className="px-4 py-3 text-sm text-slate-500">{row.database || '—'}</td>
      <td className="px-4 py-3 text-sm text-slate-500">{row.schema || '—'}</td>
      <td className="px-4 py-3 text-sm font-medium text-slate-900">{row.object || '—'}</td>
      <td className="px-4 py-3 text-sm text-slate-600">{row.index || '—'}</td>
      <td className="px-4 py-3 text-sm text-slate-700">{row.commandType || '—'}</td>
      <td className="whitespace-nowrap px-4 py-3 text-sm text-slate-600">{formatDateTime(row.startTime)}</td>
      <td className="px-4 py-3">
        <div className="flex items-center gap-2">
          <div className="h-1.5 w-16 overflow-hidden rounded-full bg-slate-100">
            <div className={`h-full rounded-full ${barColor}`} style={{ width: `${pct}%` }} />
          </div>
          <span className={`whitespace-nowrap text-sm font-medium ${row.durationSeconds >= 3600 ? 'text-red-600' : row.durationSeconds >= 900 ? 'text-amber-600' : 'text-slate-700'}`}>
            {formatDuration(row.durationSeconds)}
          </span>
        </div>
      </td>
      <td className="px-4 py-3 text-sm">
        {hasError ? (
          <span
            className="cursor-help font-semibold text-red-600"
            title={row.errorMessage ?? undefined}
          >
            #{row.errorNumber}
          </span>
        ) : (
          <span className="text-emerald-600">OK</span>
        )}
      </td>
    </tr>
  )
}

// ── Page ──────────────────────────────────────────────────────────────────────

export function LongRunningPage() {
  const [database, setDatabase]         = useState('')
  const [dateFrom, setDateFrom]         = useState('')
  const [dateTo, setDateTo]             = useState('')
  const [commandType, setCommandType]   = useState('')
  const [granularity, setGranularity]   = useState<Granularity>('day')
  const [minDuration, setMinDuration]   = useState(300)
  const [limit, setLimit]               = useState(100)

  const queryString = useMemo(() => {
    const p = new URLSearchParams()
    if (database)    p.set('database', database)
    if (dateFrom)    p.set('dateFrom', dateFrom)
    if (dateTo)      p.set('dateTo', dateTo)
    if (commandType) p.set('commandType', commandType)
    p.set('granularity', granularity)
    p.set('minDuration', String(minDuration))
    p.set('limit', String(limit))
    return `?${p.toString()}`
  }, [database, dateFrom, dateTo, commandType, granularity, minDuration, limit])

  const filtersQuery = useQuery({ queryKey: ['filters'], queryFn: getFilterOptions })
  const reportQuery  = useQuery({
    queryKey: ['long-running', database, dateFrom, dateTo, commandType, granularity, minDuration, limit],
    queryFn: () => getLongRunningReport(queryString),
  })

  const { chartData, commandTypes } = useMemo(
    () => buildChartData(reportQuery.data?.data.timeline ?? []),
    [reportQuery.data],
  )

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold">Long Running Operations</h2>
        <p className="text-slate-600">
          Maintenance operations that exceeded the minimum duration threshold, ordered by elapsed time.
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
          <span className="text-sm font-medium">Min duration</span>
          <select value={minDuration} onChange={(e) => setMinDuration(Number(e.target.value))}
            className="w-full rounded border px-3 py-2 text-sm">
            {MIN_DURATION_OPTIONS.map(({ label, value }) => (
              <option key={value} value={value}>{label}</option>
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
          Loading long running operations…
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
              Max duration per period (minutes)
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

          {/* ── Operations table ─────────────────────────────────────────────── */}
          <div className="overflow-hidden rounded-lg bg-white shadow-sm ring-1 ring-slate-200">
            <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
              <h3 className="text-sm font-semibold text-slate-700">Top slowest operations</h3>
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
              </label>
            </div>
            {reportQuery.data.data.operations.length === 0 ? (
              <div className="p-8 text-center text-slate-500">
                No operations exceeded the minimum duration threshold.
              </div>
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
                      <Th>Duration</Th>
                      <Th>Status</Th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {reportQuery.data.data.operations.map((row, i) => (
                      <OpRow
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
              {reportQuery.data.data.operations.length} operations · threshold: {formatDuration(minDuration)}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
