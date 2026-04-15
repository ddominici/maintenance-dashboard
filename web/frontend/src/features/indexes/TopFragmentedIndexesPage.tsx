import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getTopFragmentedIndexes } from '../../api/indexes'
import { getFilterOptions } from '../../api/dashboard'
import type { IndexRow } from '../../types/api'

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

function Th({ children }: { children: React.ReactNode }) {
  return (
    <th className="whitespace-nowrap px-4 py-3 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">
      {children}
    </th>
  )
}

function ThRight({ children }: { children: React.ReactNode }) {
  return (
    <th className="whitespace-nowrap px-4 py-3 text-right text-xs font-semibold uppercase tracking-wide text-slate-500">
      {children}
    </th>
  )
}

function Badge({ count, color }: { count: number; color: string }) {
  if (count === 0) return <span className="text-slate-400">0</span>
  return <span className={`font-semibold ${color}`}>{count.toLocaleString()}</span>
}

// Fragmentation thresholds mirror Ola Hallengren defaults:
//   < 5%  → negligible (slate)
//   5–30% → reorganize territory (amber)
//   > 30% → rebuild territory (red)
function fragColor(pct: number): string {
  if (pct >= 30) return 'bg-red-500'
  if (pct >= 5)  return 'bg-amber-400'
  return 'bg-emerald-400'
}

function FragCell({ value }: { value: number | null }) {
  if (value === null) return <span className="text-slate-300">—</span>
  const clamped = Math.min(value, 100)
  return (
    <div className="flex items-center gap-2">
      <div className="h-1.5 w-16 overflow-hidden rounded-full bg-slate-100">
        <div className={`h-full rounded-full ${fragColor(value)}`} style={{ width: `${clamped}%` }} />
      </div>
      <span className={`text-sm font-medium ${value >= 30 ? 'text-red-600' : value >= 5 ? 'text-amber-600' : 'text-emerald-600'}`}>
        {value.toFixed(1)}%
      </span>
    </div>
  )
}

function Row({ row, index }: { row: IndexRow; index: number }) {
  const base = index % 2 === 0 ? 'bg-white' : 'bg-slate-50'
  return (
    <tr className={`${base} hover:bg-blue-50 transition-colors`}>
      <td className="px-4 py-3 text-sm text-slate-500">{row.database || '—'}</td>
      <td className="px-4 py-3 text-sm text-slate-500">{row.schema || '—'}</td>
      <td className="px-4 py-3 text-sm font-medium text-slate-900">{row.object || '—'}</td>
      <td className="px-4 py-3 text-sm text-slate-700">{row.index || '—'}</td>
      <td className="px-4 py-3 text-right text-sm">
        <Badge count={row.rebuildCount} color="text-amber-600" />
      </td>
      <td className="px-4 py-3 text-right text-sm">
        <Badge count={row.reorganizeCount} color="text-blue-600" />
      </td>
      <td className="px-4 py-3 text-right text-sm font-semibold text-slate-900">
        {row.totalOperations.toLocaleString()}
      </td>
      <td className="whitespace-nowrap px-4 py-3 text-sm text-slate-600">{formatDate(row.lastOperation)}</td>
      <td className="px-4 py-3 text-right text-sm text-slate-600">{row.avgDurationSeconds.toFixed(2)}</td>
      <td className="px-4 py-3 text-right text-sm">
        {row.errorCount > 0 ? (
          <span className="font-semibold text-red-600">{row.errorCount}</span>
        ) : (
          <span className="text-slate-400">0</span>
        )}
      </td>
      <td className="px-4 py-3"><FragCell value={row.lastFragmentation} /></td>
      <td className="px-4 py-3"><FragCell value={row.maxFragmentation} /></td>
    </tr>
  )
}

export function TopFragmentedIndexesPage() {
  const [database, setDatabase] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [limit, setLimit] = useState(50)

  const queryString = useMemo(() => {
    const p = new URLSearchParams()
    if (database) p.set('database', database)
    if (dateFrom) p.set('dateFrom', dateFrom)
    if (dateTo) p.set('dateTo', dateTo)
    p.set('limit', String(limit))
    return `?${p.toString()}`
  }, [database, dateFrom, dateTo, limit])

  const filtersQuery = useQuery({ queryKey: ['filters'], queryFn: getFilterOptions })
  const indexesQuery = useQuery({
    queryKey: ['indexes-top-fragmented', database, dateFrom, dateTo, limit],
    queryFn: () => getTopFragmentedIndexes(queryString),
  })

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold">Top Fragmented Indexes</h2>
        <p className="text-slate-600">
          Indexes ordered by total maintenance operations (
          <span className="font-medium text-amber-600">REBUILD</span> +{' '}
          <span className="font-medium text-blue-600">REORGANIZE</span>). A high count
          indicates frequent fragmentation.
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
      {indexesQuery.isLoading && (
        <div className="rounded-lg bg-white p-8 text-center text-slate-500 shadow-sm ring-1 ring-slate-200">
          Loading indexes…
        </div>
      )}

      {/* Error */}
      {indexesQuery.isError && (
        <div className="rounded border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {(indexesQuery.error as Error).message}
        </div>
      )}

      {/* Table */}
      {indexesQuery.data && (
        <div className="overflow-hidden rounded-lg bg-white shadow-sm ring-1 ring-slate-200">
          {indexesQuery.data.data.length === 0 ? (
            <div className="p-8 text-center text-slate-500">No index operations found for the selected filters.</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-slate-200">
                <thead className="bg-slate-50">
                  <tr>
                    <Th>Database</Th>
                    <Th>Schema</Th>
                    <Th>Table</Th>
                    <Th>Index</Th>
                    <ThRight>Rebuilds</ThRight>
                    <ThRight>Reorganizes</ThRight>
                    <ThRight>Total</ThRight>
                    <Th>Last Operation</Th>
                    <ThRight>Avg Duration (s)</ThRight>
                    <ThRight>Errors</ThRight>
                    <Th>Last Fragm.</Th>
                    <Th>Max Fragm.</Th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {indexesQuery.data.data.map((row, i) => (
                    <Row
                      key={`${row.database}-${row.schema}-${row.object}-${row.index}`}
                      row={row}
                      index={i}
                    />
                  ))}
                </tbody>
              </table>
            </div>
          )}
          <div className="border-t border-slate-100 px-4 py-2 text-right text-xs text-slate-400">
            {indexesQuery.data.data.length} rows
          </div>
        </div>
      )}
    </div>
  )
}
