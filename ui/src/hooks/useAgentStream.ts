import { useEffect, useRef, useState, useCallback } from 'react'
import { useWebSocket } from '../context/WebSocketProvider'

export interface TranscriptEntry {
  timestamp: string
  stream: 'stdout' | 'stderr'
  data: string
}

interface RunLogPayload {
  runId: string
  timestamp: string
  stream: 'stdout' | 'stderr'
  data: string
}

export function useAgentStream(runId: string) {
  const { ws } = useWebSocket()
  const [transcript, setTranscript] = useState<TranscriptEntry[]>([])
  const runIdRef = useRef(runId)
  runIdRef.current = runId

  useEffect(() => {
    const unsub = ws.on('heartbeat.run.log', (payload: RunLogPayload) => {
      if (payload.runId !== runIdRef.current) return
      setTranscript((prev) => [
        ...prev,
        { timestamp: payload.timestamp, stream: payload.stream, data: payload.data },
      ])
    })
    return unsub
  }, [ws])

  const writeStdin = useCallback(
    (data: string) => {
      ws.writeStdin(runIdRef.current, data)
    },
    [ws]
  )

  return { transcript, writeStdin }
}
