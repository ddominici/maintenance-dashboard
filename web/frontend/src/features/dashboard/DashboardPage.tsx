import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getDashboardSummary, getFilterOptions } from '../../api/dashboard'

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="rounded-lg bg-white p-5 shadow-sm ring-1 ring-slate-200">
      <div className="text-sm text-slate-500">{label}</div>
      <div className="mt-2 text-3xl font-semibold">{value}</div>
    </div>
  )
}

export function DashboardPage() {
  const [database, setDatabase] = useState('')
  const [commandType, setCommandType] = useState('')
  const [onlyErrors, setOnlyErrors] = useState(false)

  const filterQuery = useMemo(() => {
    const p = new URLSearchParams()
    if (database) p.set('database', database)
    if (commandType) p.set('commandType', commandType)
    if (onlyErrors) p.set('onlyErrors', 'true')
    const s = p.toString()
    return s ? `?${s}` : ''
  }, [database, commandType, onlyErrors])

  const filtersQuery = useQuery({ queryKey: ['filters'], queryFn: getFilterOptions })
  const summaryQuery = useQuery({ queryKey: ['dashboard-summary', filterQuery], queryFn: () => getDashboardSummary(filterQuery) })

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-semibold">Dashboard</h2>
        <p className="text-slate-600">Initial vertical slice wired to live backend endpoints.</p>
      </div>

      <div className="grid grid-cols-1 gap-4 rounded-lg bg-white p-4 shadow-sm ring-1 ring-slate-200 md:grid-cols-3">
        <label className="space-y-1">
          <span className="text-sm font-medium">Database</span>
          <select value={database} onChange={(e) => setDatabase(e.target.value)} className="w-full rounded border px-3 py-2">
            <option value="">All</option>
            {filtersQuery.data?.data.databases.map((x) => <option key={x} value={x}>{x}</option>)}
          </select>
        </label>
        <label className="space-y-1">
          <span className="text-sm font-medium">Command type</span>
          <select value={commandType} onChange={(e) => setCommandType(e.target.value)} className="w-full rounded border px-3 py-2">
            <option value="">All</option>
            {filtersQuery.data?.data.commandTypes.map((x) => <option key={x} value={x}>{x}</option>)}
          </select>
        </label>
        <label className="flex items-end gap-2 pb-2">
          <input type="checkbox" checked={onlyErrors} onChange={(e) => setOnlyErrors(e.target.checked)} />
          <span className="text-sm font-medium">Only errors</span>
        </label>
      </div>

      {summaryQuery.isLoading ? <div>Loading summary...</div> : null}
      {summaryQuery.isError ? <div className="rounded border border-red-200 bg-red-50 p-4 text-red-700">{(summaryQuery.error as Error).message}</div> : null}

      {summaryQuery.data ? (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          <StatCard label="Total commands" value={summaryQuery.data.data.totalCommands} />
          <StatCard label="Total errors" value={summaryQuery.data.data.totalErrors} />
          <StatCard label="Avg duration (sec)" value={(summaryQuery.data.data.avgDurationSeconds ?? 0).toFixed(2)} />
          <StatCard label="Index rebuilds" value={summaryQuery.data.data.totalIndexRebuilds} />
          <StatCard label="Index reorganizes" value={summaryQuery.data.data.totalIndexReorganizes} />
          <StatCard label="Statistics updates" value={summaryQuery.data.data.totalStatisticsUpdates} />
        </div>
      ) : null}
    </div>
  )
}
