let currentServer = ''

export function setCurrentServer(server: string) {
  currentServer = server
}

export async function getJson<T>(url: string): Promise<T> {
  const fullUrl = currentServer
    ? url + (url.includes('?') ? '&' : '?') + `server=${encodeURIComponent(currentServer)}`
    : url

  const response = await fetch(fullUrl, {
    headers: { Accept: 'application/json' },
  })

  if (!response.ok) {
    const body = await response.text()
    throw new Error(body || `Request failed with status ${response.status}`)
  }

  return response.json() as Promise<T>
}
