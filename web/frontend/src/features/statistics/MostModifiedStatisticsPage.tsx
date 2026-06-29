import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getMostModifiedStatistics } from '../../api/statistics'
import { getFilterOptions } from '../../api/dashboard'
import type { StatisticsRow } from '../../types/api'

const LIMIT_OPTIONS = [10, 25, 50, 100, 200]

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString(undefined, {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function Th({ children, right }: { children: React.ReactNode; right?: boolean }) {
  return (
    <th className={`whitespace-nowrap px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500 ${right ? 'text-right' : 'text-left'}`}>
      {children}
    </th>
  )
}

function SortableTh({
  children,
  active,
  dir,
  onClick,
}: {
  children: React.ReactNode
  active: boolean
  dir: 'asc' | 'desc'
  onClick: () => void
}) {
  return (
    <th
      onClick={onClick}
      className="cursor-pointer select-none whitespace-nowrap px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500 hover:text-slate-700"
    >
      <span className="inline-flex items-center gap-1">
        {children}
        <span className={`text-[10px] ${active ? 'text-slate-700' : 'text-slate-300'}`}>
          {active ? (dir === 'asc' ? '▲' : '▼') : '↕'}
        </span>
      </span>
    </th>
  )
}

// Modification counter relative to page count: indicates how stale stats were
// when last updated.  Thresholds mirror the auto-update trigger (~20% of rows).
//   < 10%  → low churn (emerald)
//   10–30% → moderate (amber)
//   > 30%  → high churn (red)
function modColor(ratio: number): string {
  if (ratio >= 30) return 'bg-red-500'
  if (ratio >= 10) return 'bg-amber-400'
  return 'bg-emerald-400'
}

function ModCell({ modCounter, pageCount }: { modCounter: number | null; pageCount: number | null }) {
  if (modCounter === null) return <span className="text-slate-300">—</span>
  const formatted = modCounter.toLocaleString()
  if (pageCount === null || pageCount === 0) {
    return <span className="text-sm text-slate-600">{formatted}</span>
  }
  const ratio = (modCounter / pageCount) * 100
  const clamped = Math.min(ratio, 100)
  return (
    <div className="flex items-center gap-2">
      <div className="h-1.5 w-16 overflow-hidden rounded-full bg-slate-100">
        <div className={`h-full rounded-full ${modColor(ratio)}`} style={{ width: `${clamped}%` }} />
      </div>
      <span className={`text-sm font-medium ${ratio >= 30 ? 'text-red-600' : ratio >= 10 ? 'text-amber-600' : 'text-emerald-600'}`}>
        {formatted}
      </span>
    </div>
  )
}

function TableRow({ row, index }: { row: StatisticsRow; index: number }) {
  const base = index % 2 === 0 ? 'bg-white' : 'bg-slate-50'
  return (
    <tr className={`${base} hover:bg-blue-50 transition-colors`}>
      <td className="px-4 py-3 text-sm text-slate-700">{row.database || '—'}</td>
      <td className="px-4 py-3 text-sm text-slate-500">{row.schema || '—'}</td>
      <td className="px-4 py-3 text-sm font-medium text-slate-900">{row.object || '—'}</td>
      <td className="px-4 py-3 text-right text-sm font-semibold text-slate-900">{row.updateCount.toLocaleString()}</td>
      <td className="whitespace-nowrap px-4 py-3 text-sm text-slate-600">{formatDate(row.lastUpdated)}</td>
      <td className="px-4 py-3 text-right text-sm text-slate-600">{row.avgDurationSeconds.toFixed(2)}</td>
      <td className="px-4 py-3 text-right text-sm">
        {row.errorCount > 0 ? (
          <span className="font-semibold text-red-600">{row.errorCount}</span>
        ) : (
          <span className="text-slate-400">0</span>
        )}
      </td>
      <td className="px-4 py-3 text-right text-sm text-slate-600">
        {row.lastPageCount !== null ? row.lastPageCount.toLocaleString() : <span className="text-slate-300">—</span>}
      </td>
      <td className="px-4 py-3">
        <ModCell modCounter={row.lastModificationCounter} pageCount={row.lastPageCount} />
      </td>
      <td className="px-4 py-3 text-right text-sm text-slate-600">
        {row.maxModificationCounter !== null ? row.maxModificationCounter.toLocaleString() : <span className="text-slate-300">—</span>}
      </td>
    </tr>
  )
}

type SortKey = 'lastModificationCounter'

export function MostModifiedStatisticsPage() {
  const [database, setDatabase] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [limit, setLimit] = useState(50)
  const [sortKey, setSortKey] = useState<SortKey | null>(null)
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')

  function toggleSort(key: SortKey) {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'))
    } else {
      setSortKey(key)
      setSortDir('desc')
    }
  }

  const queryString = useMemo(() => {
    const p = new URLSearchParams()
    if (database) p.set('database', database)
    if (dateFrom) p.set('dateFrom', dateFrom)
    if (dateTo) p.set('dateTo', dateTo)
    p.set('limit', String(limit))
    // Sort server-side so the limit captures the true top rows of the full
    // dataset, not just a re-ordering of the default top-N.
    if (sortKey) {
      p.set('sortBy', sortKey)
      p.set('sortDir', sortDir)
    }
    return `?${p.toString()}`
  }, [database, dateFrom, dateTo, limit, sortKey, sortDir])

  const filtersQuery = useQuery({ queryKey: ['filters'], queryFn: getFilterOptions })
  const statsQuery = useQuery({
    queryKey: ['statistics-most-modified', database, dateFrom, dateTo, limit, sortKey, sortDir],
    queryFn: () => getMostModifiedStatistics(queryString),
  })

  const rows = statsQuery.data?.data ?? []

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold">Most Modified Statistics</h2>
        <p className="text-slate-600">
          Tables whose statistics have been updated most often via <code className="rounded bg-slate-100 px-1 py-0.5 text-xs">UPDATE_STATISTICS</code>.
        </p>
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
          <span className="text-sm font-medium">Top</span>
          <select
            value={limit}
            onChange={(e) => setLimit(Number(e.target.value))}
            className="w-full rounded border px-3 py-2 text-sm"
          >
            {LIMIT_OPTIONS.map((n) => (
              <option key={n} value={n}>Top {n}</option>
            ))}
          </select>
        </label>
      </div>

      {/* Loading */}
      {statsQuery.isLoading && (
        <div className="rounded-lg bg-white p-8 text-center text-slate-500 shadow-sm ring-1 ring-slate-200">
          Loading statistics…
        </div>
      )}

      {/* Error */}
      {statsQuery.isError && (
        <div className="rounded border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {(statsQuery.error as Error).message}
        </div>
      )}

      {/* Table */}
      {statsQuery.data && (
        <div className="overflow-hidden rounded-lg bg-white shadow-sm ring-1 ring-slate-200">
          {statsQuery.data.data.length === 0 ? (
            <div className="p-8 text-center text-slate-500">No statistics updates found for the selected filters.</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-slate-200">
                <thead className="bg-slate-50">
                  <tr>
                    <Th>Database</Th>
                    <Th>Schema</Th>
                    <Th>Object</Th>
                    <Th right>Updates</Th>
                    <Th>Last Updated</Th>
                    <Th right>Avg Duration (s)</Th>
                    <Th right>Errors</Th>
                    <Th right>Page Count</Th>
                    <SortableTh
                      active={sortKey === 'lastModificationCounter'}
                      dir={sortDir}
                      onClick={() => toggleSort('lastModificationCounter')}
                    >
                      Last Mod. Counter
                    </SortableTh>
                    <Th right>Max Mod. Counter</Th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {rows.map((row, i) => (
                    <TableRow key={`${row.database}-${row.schema}-${row.object}`} row={row} index={i} />
                  ))}
                </tbody>
              </table>
            </div>
          )}
          <div className="border-t border-slate-100 px-4 py-2 text-right text-xs text-slate-400">
            {statsQuery.data.data.length} rows
          </div>
        </div>
      )}
    </div>
  )
}
