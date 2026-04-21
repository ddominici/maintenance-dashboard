import { createContext, useContext, useEffect, useState, PropsWithChildren } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { getServers, getServerStatus, ServerStatus } from '../api/meta'
import { setCurrentServer } from '../api/client'

interface ServerContextValue {
  servers: string[]
  serverStatuses: ServerStatus[]
  selectedServer: string
  setSelectedServer: (server: string) => void
  isLoading: boolean
  serverError: string | null
}

const ServerContext = createContext<ServerContextValue>({
  servers: [],
  serverStatuses: [],
  selectedServer: '',
  setSelectedServer: () => {},
  isLoading: true,
  serverError: null,
})

export function useServer() {
  return useContext(ServerContext)
}

export function ServerProvider({ children }: PropsWithChildren) {
  const [servers, setServers] = useState<string[]>([])
  const [serverStatuses, setServerStatuses] = useState<ServerStatus[]>([])
  const [selectedServer, setSelectedServerState] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [serverError, setServerError] = useState<string | null>(null)
  const queryClient = useQueryClient()

  useEffect(() => {
    Promise.all([getServers(), getServerStatus()])
      .then(([serversRes, statusRes]) => {
        const list = serversRes.data
        const statuses = statusRes.data
        setServers(list)
        setServerStatuses(statuses)
        setServerError(null)
        const stored = localStorage.getItem('selectedServer')
        const initial = stored && list.includes(stored) ? stored : (list[0] ?? '')
        setCurrentServer(initial)
        setSelectedServerState(initial)
      })
      .catch((err: unknown) => {
        const msg = err instanceof Error ? err.message : 'Server non raggiungibile'
        setServerError(msg)
      })
      .finally(() => setIsLoading(false))
  }, [])

  function setSelectedServer(server: string) {
    setCurrentServer(server)
    setSelectedServerState(server)
    localStorage.setItem('selectedServer', server)
    queryClient.invalidateQueries()
  }

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-100 text-slate-500">
        Connecting…
      </div>
    )
  }

  return (
    <ServerContext.Provider value={{ servers, serverStatuses, selectedServer, setSelectedServer, isLoading, serverError }}>
      {children}
    </ServerContext.Provider>
  )
}
