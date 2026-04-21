import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { getServerStatus } from '../../api/meta'
import { useServer } from '../../app/ServerContext'

function StatusBadge({ reachable }: { reachable: boolean }) {
  if (reachable) {
    return (
      <span className="inline-flex items-center gap-1.5 rounded-full bg-emerald-100 px-2.5 py-0.5 text-xs font-medium text-emerald-700">
        <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
        Online
      </span>
    )
  }
  return (
    <span className="inline-flex items-center gap-1.5 rounded-full bg-red-100 px-2.5 py-0.5 text-xs font-medium text-red-700">
      <span className="h-1.5 w-1.5 rounded-full bg-red-500" />
      Offline
    </span>
  )
}

export function ServersPage() {
  const navigate = useNavigate()
  const { setSelectedServer, selectedServer } = useServer()

  const { data, isLoading, error, refetch, isFetching } = useQuery({
    queryKey: ['server-status'],
    queryFn: getServerStatus,
    refetchInterval: 30_000,
  })

  const servers = data?.data ?? []

  function handleConnect(name: string) {
    setSelectedServer(name)
    navigate('/')
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold">Servers</h2>
          <p className="text-slate-500">Elenco dei server monitorati e relativo stato di raggiungibilità.</p>
        </div>
        <button
          onClick={() => refetch()}
          disabled={isFetching}
          className="flex items-center gap-1.5 rounded border border-slate-300 bg-white px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-50 disabled:opacity-50"
        >
          <svg className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Aggiorna
        </button>
      </div>

      {error && (
        <div className="rounded-lg bg-red-50 px-4 py-3 text-sm text-red-700 ring-1 ring-red-200">
          Impossibile recuperare lo stato dei server: {error instanceof Error ? error.message : String(error)}
        </div>
      )}

      {isLoading && (
        <div className="text-sm text-slate-500">Verifica raggiungibilità server…</div>
      )}

      {!isLoading && servers.length === 0 && !error && (
        <div className="text-sm text-slate-500">Nessun server configurato.</div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {servers.map((server) => {
          const isSelected = server.name === selectedServer
          return (
            <div
              key={server.name}
              className={`rounded-lg bg-white p-5 shadow-sm ring-1 transition-shadow ${
                isSelected ? 'ring-slate-500 shadow-md' : 'ring-slate-200 hover:shadow-md'
              }`}
            >
              <div className="flex items-start justify-between gap-2">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="truncate font-semibold text-slate-800">{server.name}</span>
                    {isSelected && (
                      <span className="shrink-0 rounded bg-slate-800 px-1.5 py-0.5 text-[10px] font-medium text-white">
                        Attivo
                      </span>
                    )}
                  </div>
                  <div className="mt-1 font-mono text-xs text-slate-500">{server.host}</div>
                </div>
                <StatusBadge reachable={server.reachable} />
              </div>

              <div className="mt-4">
                {server.reachable ? (
                  <button
                    onClick={() => handleConnect(server.name)}
                    disabled={isSelected}
                    className="w-full rounded bg-slate-800 px-3 py-1.5 text-sm text-white hover:bg-slate-700 disabled:cursor-default disabled:bg-slate-300 disabled:text-slate-500"
                  >
                    {isSelected ? 'Connesso' : 'Connetti'}
                  </button>
                ) : (
                  <button
                    disabled
                    className="w-full cursor-not-allowed rounded bg-slate-100 px-3 py-1.5 text-sm text-slate-400"
                  >
                    Non raggiungibile
                  </button>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
