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

function ServerUnavailable() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 p-6 text-center">
      <div className="rounded-full bg-red-100 p-4">
        <svg className="h-10 w-10 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
        </svg>
      </div>
      <h2 className="text-xl font-semibold text-slate-700">Server non disponibile</h2>
      <p className="max-w-sm text-slate-500">
        Impossibile connettersi al server backend. Verificare che il server sia in esecuzione e ricaricare la pagina.
      </p>
      <button
        onClick={() => window.location.reload()}
        className="rounded bg-slate-800 px-4 py-2 text-sm text-white hover:bg-slate-700"
      >
        Riprova
      </button>
    </div>
  )
}

// Guard: redirects to /servers if the selected server is unreachable
function ReachableGuard() {
  const { serverStatuses, selectedServer } = useServer()
  const status = serverStatuses.find((s) => s.name === selectedServer)
  if (status && !status.reachable) {
    return <Navigate to="/servers" replace />
  }
  return <Outlet />
}

function Layout() {
  const { servers, serverStatuses, selectedServer, setSelectedServer, serverError } = useServer()

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

          {serverError && (
            <div className="mt-auto pt-4">
              <div className="flex items-start gap-2 rounded bg-red-900/60 px-3 py-2 text-xs text-red-200 ring-1 ring-red-700">
                <svg className="mt-0.5 h-3.5 w-3.5 shrink-0 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m0 3.75h.008v.008H12v-.008Zm9.303-3.376c.866 1.5-.217 3.374-1.948 3.374H4.645c-1.73 0-2.813-1.874-1.948-3.374L10.052 3.378c.866-1.5 3.032-1.5 3.898 0l7.353 12.748Z" />
                </svg>
                <span className="leading-tight">Server non raggiungibile</span>
              </div>
            </div>
          )}
        </aside>
        <main className="flex flex-1 flex-col p-6">
          {serverError ? (
            <ServerUnavailable />
          ) : (
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
          )}
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
