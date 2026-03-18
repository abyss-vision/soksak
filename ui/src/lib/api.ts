export function getBaseUrl(): string {
  if (typeof window !== 'undefined') {
    // Electron exposes server URL via preload
    const electronUrl = (window as any).electronAPI?.getServerUrl?.()
    if (electronUrl) return electronUrl
  }
  if (import.meta.env.VITE_API_URL) return import.meta.env.VITE_API_URL
  return window.location.origin
}

export function getWsUrl(): string {
  const base = getBaseUrl()
  // Replace http(s) with ws(s)
  return base.replace(/^http/, 'ws')
}
