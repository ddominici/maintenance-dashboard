import { BrowserRouter, Link, Navigate, Outlet, Route, Routes } from 'react-router-dom'
import { DashboardPage } from '../features/dashboard/DashboardPage'
import { MostModifiedStatisticsPage } from '../features/statistics/MostModifiedStatisticsPage'
import { TopFragmentedIndexesPage } from '../features/indexes/TopFragmentedIndexesPage'
import { MaintenanceSummaryPage } from '../features/maintenance/MaintenanceSummaryPage'
import { BackupPage } from '../features/backup/BackupPage'
import { OperationsPerBatchPage } from '../features/operations/OperationsPerBatchPage'
import { LongRunningPage } from '../features/longrunning/LongRunningPage'
import { MaintenanceErrorsPage } from '../features/errors/MaintenanceErrorsPage'
import { ServersPage } from '../features/servers/ServersPage'
import { useServer } from './ServerContext'

// Guard: redirects to /servers if the backend is unreachable, no servers are configured,
// or the selected server is offline — no error screen is shown.
function ReachableGuard() {
  const { serverStatuses, selectedServer, serverError, servers } = useServer()
  if (serverError || servers.length === 0) {
    return <Navigate to="/servers" replace />
  }
  const status = serverStatuses.find((s) => s.name === selectedServer)
  if (status && !status.reachable) {
    return <Navigate to="/servers" replace />
  }
  return <Outlet />
}

function Layout() {
  const { servers, serverStatuses, selectedServer, setSelectedServer } = useServer()

  const links = [
    ['/', 'Dashboard'],
    ['/fragmented-indexes', 'Top Fragmented Indexes'],
    ['/modified-statistics', 'Most Modified Statistics'],
    ['/operations-per-batch', 'Operations Per Batch'],
    ['/maintenance-summary', 'Maintenance Summary'],
    ['/backup-overview', 'Backup Overview'],
    ['/maintenance-errors', 'Maintenance Errors'],
    ['/long-running-operations', 'Long Running Operations'],
    ['/servers', 'Servers'],
  ]

  return (
    <div className="min-h-screen bg-slate-100 text-slate-900">
      <div className="flex">
        <aside className="flex min-h-screen w-72 flex-col bg-slate-900 p-4 text-white">
          <h1 className="mb-4 text-xl font-semibold">Maintenance Dashboard</h1>

          {servers.length > 0 && (
            <div className="mb-5">
              <label className="mb-1 block text-xs font-medium text-slate-400">Server</label>
              <select
                value={selectedServer}
                onChange={(e) => setSelectedServer(e.target.value)}
                className="w-full rounded border border-slate-600 bg-slate-800 px-2 py-1.5 text-sm text-white focus:outline-none focus:ring-1 focus:ring-slate-400"
              >
                {servers.map((s) => {
                  const st = serverStatuses.find((x) => x.name === s)
                  const offline = st != null && !st.reachable
                  return (
                    <option key={s} value={s} disabled={offline}>
                      {s}{offline ? ' (offline)' : ''}
                    </option>
                  )
                })}
              </select>
            </div>
          )}

          <nav className="space-y-2">
            {links.map(([to, label]) => (
              <Link key={to} to={to} className="block rounded px-3 py-2 hover:bg-slate-800">
                {label}
              </Link>
            ))}
          </nav>

        </aside>
        <main className="flex flex-1 flex-col p-6">
          <Routes>
            <Route path="/servers" element={<ServersPage />} />
            <Route element={<ReachableGuard />}>
              <Route path="/" element={<DashboardPage />} />
              <Route path="/fragmented-indexes" element={<TopFragmentedIndexesPage />} />
              <Route path="/modified-statistics" element={<MostModifiedStatisticsPage />} />
              <Route path="/operations-per-batch" element={<OperationsPerBatchPage />} />
              <Route path="/maintenance-summary" element={<MaintenanceSummaryPage />} />
              <Route path="/backup-overview" element={<BackupPage />} />
              <Route path="/maintenance-errors" element={<MaintenanceErrorsPage />} />
              <Route path="/long-running-operations" element={<LongRunningPage />} />
            </Route>
          </Routes>
        </main>
      </div>
    </div>
  )
}

export function AppRouter() {
  return (
    <BrowserRouter>
      <Layout />
    </BrowserRouter>
  )
}
