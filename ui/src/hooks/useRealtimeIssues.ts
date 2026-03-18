import { useEffect } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useWebSocket } from '../context/WebSocketProvider'
import { getBaseUrl } from '../lib/api'

export interface Issue {
  uuid: string
  title: string
  status: string
  priority: string
  [key: string]: any
}

async function fetchIssues(companyUuid: string): Promise<Issue[]> {
  const res = await fetch(`${getBaseUrl()}/api/companies/${companyUuid}/issues`)
  if (!res.ok) throw new Error('Failed to fetch issues')
  return res.json()
}

export function useRealtimeIssues(companyUuid: string) {
  const queryClient = useQueryClient()
  const { ws } = useWebSocket()

  const query = useQuery<Issue[]>({
    queryKey: ['issues', companyUuid],
    queryFn: () => fetchIssues(companyUuid),
    staleTime: Infinity,
    refetchOnWindowFocus: false,
  })

  useEffect(() => {
    const unsubUpdated = ws.on('issue.updated', (payload: Issue) => {
      queryClient.setQueryData<Issue[]>(['issues', companyUuid], (old = []) =>
        old.map((issue) => (issue.uuid === payload.uuid ? { ...issue, ...payload } : issue))
      )
    })

    const unsubCreated = ws.on('issue.created', (payload: Issue) => {
      queryClient.setQueryData<Issue[]>(['issues', companyUuid], (old = []) => [
        ...old,
        payload,
      ])
    })

    const unsubMoved = ws.on('kanban.issue.moved', (payload: { issueUuid: string; toStatus: string }) => {
      queryClient.setQueryData<Issue[]>(['issues', companyUuid], (old = []) =>
        old.map((issue) =>
          issue.uuid === payload.issueUuid ? { ...issue, status: payload.toStatus } : issue
        )
      )
    })

    return () => {
      unsubUpdated()
      unsubCreated()
      unsubMoved()
    }
  }, [ws, companyUuid, queryClient])

  return query
}
