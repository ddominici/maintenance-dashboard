import { BrowserRouter, Link, Route, Routes } from 'react-router-dom'
import { DashboardPage } from '../features/dashboard/DashboardPage'
import { MostModifiedStatisticsPage } from '../features/statistics/MostModifiedStatisticsPage'
import { TopFragmentedIndexesPage } from '../features/indexes/TopFragmentedIndexesPage'
import { MaintenanceSummaryPage } from '../features/maintenance/MaintenanceSummaryPage'
import { BackupPage } from '../features/backup/BackupPage'
import { OperationsPerBatchPage } from '../features/operations/OperationsPerBatchPage'
import { LongRunningPage } from '../features/longrunning/LongRunningPage'
import { MaintenanceErrorsPage } from '../features/errors/MaintenanceErrorsPage'

function Placeholder({ title }: { title: string }) {
  return <div className="rounded border bg-white p-6 shadow-sm">{title} - coming next</div>
}

function Layout() {
  const links = [
    ['/', 'Dashboard'],
    ['/fragmented-indexes', 'Top Fragmented Indexes'],
    ['/modified-statistics', 'Most Modified Statistics'],
    ['/operations-per-batch', 'Operations Per Batch'],
    ['/maintenance-summary', 'Maintenance Summary'],
    ['/backup-overview', 'Backup Overview'],
    ['/maintenance-errors', 'Maintenance Errors'],
    ['/long-running-operations', 'Long Running Operations'],
  ]
  return (
    <div className="min-h-screen bg-slate-100 text-slate-900">
      <div className="flex">
        <aside className="min-h-screen w-72 bg-slate-900 p-4 text-white">
          <h1 className="mb-4 text-xl font-semibold">Maintenance Dashboard</h1>
          <nav className="space-y-2">
            {links.map(([to, label]) => (
              <Link key={to} to={to} className="block rounded px-3 py-2 hover:bg-slate-800">
                {label}
              </Link>
            ))}
          </nav>
        </aside>
        <main className="flex-1 p-6">
          <Routes>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/fragmented-indexes" element={<TopFragmentedIndexesPage />} />
            <Route path="/modified-statistics" element={<MostModifiedStatisticsPage />} />
            <Route path="/operations-per-batch" element={<OperationsPerBatchPage />} />
            <Route path="/maintenance-summary" element={<MaintenanceSummaryPage />} />
            <Route path="/backup-overview" element={<BackupPage />} />
            <Route path="/maintenance-errors" element={<MaintenanceErrorsPage />} />
            <Route path="/long-running-operations" element={<LongRunningPage />} />
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
