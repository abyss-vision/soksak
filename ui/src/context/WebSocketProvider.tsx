import {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  ReactNode,
} from 'react'
import { getWsUrl } from '../lib/api'

type MessageHandler = (payload: any) => void

interface Subscription {
  type: string
  handler: MessageHandler
}

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected'

class SoksakWebSocket {
  private ws: WebSocket | null = null
  private companyUuid: string = ''
  private subscriptions: Set<Subscription> = new Set()
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectDelay = 1000
  private maxDelay = 30000
  private destroyed = false
  private statusListeners: Set<(status: ConnectionStatus) => void> = new Set()

  connect(companyUuid: string): void {
    this.companyUuid = companyUuid
    this.destroyed = false
    this.openSocket()
  }

  private openSocket(): void {
    if (this.destroyed) return
    const url = `${getWsUrl()}/api/companies/${this.companyUuid}/events/ws`
    this.ws = new WebSocket(url)
    this.notifyStatus('connecting')

    this.ws.onopen = () => {
      this.reconnectDelay = 1000
      this.notifyStatus('connected')
    }

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data as string)
        const { type, payload } = msg
        this.subscriptions.forEach((sub) => {
          if (sub.type === type) {
            sub.handler(payload)
          }
        })
      } catch {
        // ignore malformed messages
      }
    }

    this.ws.onclose = () => {
      this.notifyStatus('disconnected')
      if (!this.destroyed) {
        this.scheduleReconnect()
      }
    }

    this.ws.onerror = () => {
      // onclose will fire after onerror; reconnect handled there
    }
  }

  private scheduleReconnect(): void {
    this.reconnectTimer = setTimeout(() => {
      this.openSocket()
    }, this.reconnectDelay)
    this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxDelay)
  }

  send(type: string, payload: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, payload }))
    }
  }

  on(type: string, handler: MessageHandler): () => void {
    const sub: Subscription = { type, handler }
    this.subscriptions.add(sub)
    return () => {
      this.subscriptions.delete(sub)
    }
  }

  onStatus(listener: (status: ConnectionStatus) => void): () => void {
    this.statusListeners.add(listener)
    return () => {
      this.statusListeners.delete(listener)
    }
  }

  private notifyStatus(status: ConnectionStatus): void {
    this.statusListeners.forEach((l) => l(status))
  }

  // Convenience helpers

  writeStdin(runId: string, data: string): void {
    this.send('agent.stdin', { runId, data })
  }

  signalAgent(runId: string, signal: string): void {
    this.send('agent.signal', { runId, signal })
  }

  moveIssue(issueUuid: string, toStatus: string): void {
    this.send('kanban.move', { issueUuid, toStatus })
  }

  destroy(): void {
    this.destroyed = true
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.ws?.close()
    this.ws = null
  }
}

interface WebSocketContextValue {
  ws: SoksakWebSocket
  status: ConnectionStatus
}

const WebSocketContext = createContext<WebSocketContextValue | null>(null)

interface WebSocketProviderProps {
  companyUuid: string
  children: ReactNode
}

export function WebSocketProvider({ companyUuid, children }: WebSocketProviderProps) {
  const wsRef = useRef<SoksakWebSocket>(new SoksakWebSocket())
  const [status, setStatus] = useState<ConnectionStatus>('disconnected')

  useEffect(() => {
    const ws = wsRef.current
    const unsub = ws.onStatus(setStatus)
    ws.connect(companyUuid)
    return () => {
      unsub()
      ws.destroy()
    }
  }, [companyUuid])

  return (
    <WebSocketContext.Provider value={{ ws: wsRef.current, status }}>
      {children}
    </WebSocketContext.Provider>
  )
}

export function useWebSocket(): WebSocketContextValue {
  const ctx = useContext(WebSocketContext)
  if (!ctx) throw new Error('useWebSocket must be used inside WebSocketProvider')
  return ctx
}

export type { SoksakWebSocket }
