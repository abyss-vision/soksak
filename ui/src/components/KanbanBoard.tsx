import { useState, useCallback, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import {
  DndContext,
  DragEndEvent,
  DragOverEvent,
  DragStartEvent,
  PointerSensor,
  useSensor,
  useSensors,
  DragOverlay,
  useDroppable,
  useDraggable,
} from '@dnd-kit/core'
import { useRealtimeIssues, Issue } from '../hooks/useRealtimeIssues'
import { useWebSocket, ConnectionStatus } from '../context/WebSocketProvider'
import { getBaseUrl } from '../lib/api'

// ---------------------------------------------------------------------------
// Status machine — mirrors Go state machine
// ---------------------------------------------------------------------------
type IssueStatus = 'backlog' | 'todo' | 'in_progress' | 'in_review' | 'done' | 'cancelled'

const VALID_TRANSITIONS: Record<IssueStatus, IssueStatus[]> = {
  backlog: ['todo', 'cancelled'],
  todo: ['in_progress', 'backlog', 'cancelled'],
  in_progress: ['in_review', 'todo', 'cancelled'],
  in_review: ['done', 'in_progress', 'cancelled'],
  done: [],
  cancelled: ['backlog'],
}

export function isValidDrop(from: IssueStatus, to: IssueStatus): boolean {
  if (from === to) return false
  return VALID_TRANSITIONS[from]?.includes(to) ?? false
}

// ---------------------------------------------------------------------------
// Column config
// ---------------------------------------------------------------------------
const DEFAULT_COLUMN_ORDER: IssueStatus[] = [
  'backlog',
  'todo',
  'in_progress',
  'in_review',
  'done',
]

const COLUMN_LABEL_KEYS: Record<IssueStatus, string> = {
  backlog: 'status.backlog',
  todo: 'status.todo',
  in_progress: 'status.in_progress',
  in_review: 'status.in_review',
  done: 'status.done',
  cancelled: 'status.cancelled',
}

const COLLAPSED_KEY = 'kanban-collapsed'

function loadCollapsed(): Set<string> {
  try {
    const raw = localStorage.getItem(COLLAPSED_KEY)
    if (raw) return new Set(JSON.parse(raw))
  } catch {
    // ignore
  }
  return new Set()
}

function saveCollapsed(set: Set<string>): void {
  localStorage.setItem(COLLAPSED_KEY, JSON.stringify([...set]))
}

// ---------------------------------------------------------------------------
// Board settings types
// ---------------------------------------------------------------------------
interface BoardSettingsData {
  uuid: string
  companyUuid: string
  columnOrder: IssueStatus[]
  hiddenColumns: IssueStatus[]
  swimLaneField: string | null
  createdAt: string
  updatedAt: string
}

async function fetchBoardSettings(companyUuid: string): Promise<BoardSettingsData> {
  const res = await fetch(`${getBaseUrl()}/api/companies/${companyUuid}/board-settings`)
  if (!res.ok) throw new Error('Failed to fetch board settings')
  return res.json()
}

async function patchBoardSettings(
  companyUuid: string,
  patch: Partial<Pick<BoardSettingsData, 'columnOrder' | 'hiddenColumns' | 'swimLaneField'>>
): Promise<BoardSettingsData> {
  const res = await fetch(`${getBaseUrl()}/api/companies/${companyUuid}/board-settings`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(patch),
  })
  if (!res.ok) throw new Error('Failed to update board settings')
  return res.json()
}

// ---------------------------------------------------------------------------
// Connection status badge
// ---------------------------------------------------------------------------
function ConnectionBadge({ status }: { status: ConnectionStatus }) {
  const { t } = useTranslation()
  const label =
    status === 'connected'
      ? t('status.connected', 'Connected')
      : status === 'connecting'
      ? t('status.connecting', 'Connecting…')
      : t('status.disconnected', 'Disconnected')
  const color =
    status === 'connected'
      ? 'bg-green-100 text-green-700'
      : status === 'connecting'
      ? 'bg-yellow-100 text-yellow-700'
      : 'bg-red-100 text-red-700'
  return (
    <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${color}`}>{label}</span>
  )
}

// ---------------------------------------------------------------------------
// Board settings dialog
// ---------------------------------------------------------------------------
interface BoardSettingsDialogProps {
  settings: BoardSettingsData
  onClose: () => void
  onSave: (patch: Partial<Pick<BoardSettingsData, 'columnOrder' | 'hiddenColumns' | 'swimLaneField'>>) => Promise<void>
}

function BoardSettingsDialog({ settings, onClose, onSave }: BoardSettingsDialogProps) {
  const { t } = useTranslation()
  const [order, setOrder] = useState<IssueStatus[]>(
    settings.columnOrder.length > 0 ? settings.columnOrder : DEFAULT_COLUMN_ORDER
  )
  const [hidden, setHidden] = useState<Set<IssueStatus>>(
    new Set(settings.hiddenColumns as IssueStatus[])
  )
  const [swimLane, setSwimLane] = useState<boolean>(settings.swimLaneField === 'assigneeAgentId')
  const [saving, setSaving] = useState(false)

  function toggleHidden(col: IssueStatus) {
    setHidden((prev) => {
      const next = new Set(prev)
      if (next.has(col)) next.delete(col)
      else next.add(col)
      return next
    })
  }

  function moveUp(idx: number) {
    if (idx === 0) return
    setOrder((prev) => {
      const next = [...prev]
      ;[next[idx - 1], next[idx]] = [next[idx], next[idx - 1]]
      return next
    })
  }

  function moveDown(idx: number) {
    if (idx === order.length - 1) return
    setOrder((prev) => {
      const next = [...prev]
      ;[next[idx], next[idx + 1]] = [next[idx + 1], next[idx]]
      return next
    })
  }

  async function handleSave() {
    setSaving(true)
    try {
      await onSave({
        columnOrder: order,
        hiddenColumns: [...hidden] as IssueStatus[],
        swimLaneField: swimLane ? 'assigneeAgentId' : null,
      })
      onClose()
    } finally {
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      role="dialog"
      aria-modal="true"
      aria-label={t('kanban.boardSettings', 'Board Settings')}
    >
      <div className="bg-white rounded-xl shadow-xl w-full max-w-sm p-6 flex flex-col gap-5">
        <h2 className="text-base font-semibold text-gray-800">
          {t('kanban.boardSettings', 'Board Settings')}
        </h2>

        {/* Column ordering + visibility */}
        <section>
          <p className="text-xs font-medium text-gray-500 mb-2 uppercase tracking-wide">
            {t('kanban.columnOrder', 'Column Order & Visibility')}
          </p>
          <ul className="space-y-1">
            {order.map((col, idx) => (
              <li key={col} className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id={`col-vis-${col}`}
                  checked={!hidden.has(col)}
                  onChange={() => toggleHidden(col)}
                  className="h-4 w-4 rounded border-gray-300 text-blue-600"
                  aria-label={t(COLUMN_LABEL_KEYS[col], col)}
                />
                <label
                  htmlFor={`col-vis-${col}`}
                  className="flex-1 text-sm text-gray-700 capitalize cursor-pointer"
                >
                  {t(COLUMN_LABEL_KEYS[col], col)}
                </label>
                <button
                  onClick={() => moveUp(idx)}
                  disabled={idx === 0}
                  aria-label={t('kanban.moveUp', 'Move up')}
                  className="p-0.5 text-gray-400 hover:text-gray-700 disabled:opacity-30"
                >
                  ▲
                </button>
                <button
                  onClick={() => moveDown(idx)}
                  disabled={idx === order.length - 1}
                  aria-label={t('kanban.moveDown', 'Move down')}
                  className="p-0.5 text-gray-400 hover:text-gray-700 disabled:opacity-30"
                >
                  ▼
                </button>
              </li>
            ))}
          </ul>
        </section>

        {/* Swim lane toggle */}
        <section className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-gray-700">
              {t('kanban.swimLanes', 'Swim Lanes')}
            </p>
            <p className="text-xs text-gray-400">
              {t('kanban.swimLanesDesc', 'Group issues by assignee')}
            </p>
          </div>
          <button
            role="switch"
            aria-checked={swimLane}
            onClick={() => setSwimLane((v) => !v)}
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
              swimLane ? 'bg-blue-600' : 'bg-gray-300'
            }`}
          >
            <span
              className={`inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform ${
                swimLane ? 'translate-x-6' : 'translate-x-1'
              }`}
            />
          </button>
        </section>

        {/* Actions */}
        <div className="flex justify-end gap-2 pt-1">
          <button
            onClick={onClose}
            className="rounded-md px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100"
          >
            {t('common.cancel', 'Cancel')}
          </button>
          <button
            onClick={handleSave}
            disabled={saving}
            className="rounded-md bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-60"
          >
            {saving ? t('common.saving', 'Saving…') : t('common.save', 'Save')}
          </button>
        </div>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Draggable issue card
// ---------------------------------------------------------------------------
function IssueCard({ issue, isDragging }: { issue: Issue; isDragging?: boolean }) {
  const { attributes, listeners, setNodeRef, transform } = useDraggable({ id: issue.uuid })
  const style = transform
    ? { transform: `translate(${transform.x}px, ${transform.y}px)` }
    : undefined

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...listeners}
      {...attributes}
      data-status={issue.status}
      className={`rounded-md border border-gray-200 bg-white p-3 shadow-sm cursor-grab select-none ${
        isDragging ? 'opacity-50' : 'hover:border-blue-300'
      }`}
    >
      <p className="text-sm font-medium text-gray-800 truncate">{issue.title}</p>
      {issue.priority && issue.priority !== 'none' && (
        <span className="mt-1 inline-block text-xs text-gray-400 capitalize">{issue.priority}</span>
      )}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Droppable column
// ---------------------------------------------------------------------------
function KanbanColumn({
  columnId,
  label,
  issues,
  isDimmed,
  isCollapsed,
  onToggleCollapse,
  extraHeader,
}: {
  columnId: string
  label: string
  issues: Issue[]
  isDimmed: boolean
  isCollapsed: boolean
  onToggleCollapse: () => void
  extraHeader?: React.ReactNode
}) {
  const { setNodeRef, isOver } = useDroppable({ id: columnId })

  return (
    <div
      className={`flex flex-col min-w-[220px] w-[220px] transition-opacity ${
        isDimmed ? 'opacity-30 pointer-events-none' : 'opacity-100'
      }`}
    >
      <div className="flex items-center justify-between mb-2">
        <button
          onClick={onToggleCollapse}
          className="flex items-center gap-1 text-sm font-semibold text-gray-700 hover:text-gray-900"
        >
          <span className="text-xs">{isCollapsed ? '▶' : '▼'}</span>
          {label}
          <span className="ml-1 text-xs font-normal text-gray-400">({issues.length})</span>
        </button>
        {extraHeader}
      </div>

      {!isCollapsed && (
        <div
          ref={setNodeRef}
          className={`flex-1 min-h-[120px] rounded-lg p-2 space-y-2 transition-colors ${
            isOver ? 'bg-blue-50 ring-2 ring-blue-300' : 'bg-gray-100'
          }`}
        >
          {issues.map((issue) => (
            <IssueCard key={issue.uuid} issue={issue} />
          ))}
        </div>
      )}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Done column (done + cancelled split)
// ---------------------------------------------------------------------------
function DoneColumn({
  doneIssues,
  cancelledIssues,
  isDimmedDone,
  isDimmedCancelled,
  isCollapsedDone,
  isCollapsedCancelled,
  onToggleDone,
  onToggleCancelled,
}: {
  doneIssues: Issue[]
  cancelledIssues: Issue[]
  isDimmedDone: boolean
  isDimmedCancelled: boolean
  isCollapsedDone: boolean
  isCollapsedCancelled: boolean
  onToggleDone: () => void
  onToggleCancelled: () => void
}) {
  const { t } = useTranslation()
  const { setNodeRef: doneRef, isOver: isOverDone } = useDroppable({ id: 'done' })
  const { setNodeRef: cancelledRef, isOver: isOverCancelled } = useDroppable({ id: 'cancelled' })

  return (
    <div className="flex flex-col gap-3 min-w-[220px] w-[220px]">
      {/* Done sub-column */}
      <div className={`transition-opacity ${isDimmedDone ? 'opacity-30 pointer-events-none' : ''}`}>
        <button
          onClick={onToggleDone}
          className="flex items-center gap-1 text-sm font-semibold text-green-700 hover:text-green-900 mb-1"
        >
          <span className="text-xs">{isCollapsedDone ? '▶' : '▼'}</span>
          <span className="text-green-600">✓</span>
          {t('status.done')}
          <span className="ml-1 text-xs font-normal text-gray-400">({doneIssues.length})</span>
        </button>
        {!isCollapsedDone && (
          <div
            ref={doneRef}
            className={`min-h-[60px] rounded-lg p-2 space-y-2 transition-colors ${
              isOverDone ? 'bg-green-50 ring-2 ring-green-300' : 'bg-gray-100'
            }`}
          >
            {doneIssues.map((issue) => (
              <IssueCard key={issue.uuid} issue={issue} />
            ))}
          </div>
        )}
      </div>

      {/* Cancelled sub-column */}
      <div className={`transition-opacity ${isDimmedCancelled ? 'opacity-30 pointer-events-none' : ''}`}>
        <button
          onClick={onToggleCancelled}
          className="flex items-center gap-1 text-sm font-semibold text-red-700 hover:text-red-900 mb-1"
        >
          <span className="text-xs">{isCollapsedCancelled ? '▶' : '▼'}</span>
          <span className="text-red-600">✕</span>
          {t('status.cancelled')}
          <span className="ml-1 text-xs font-normal text-gray-400">({cancelledIssues.length})</span>
        </button>
        {!isCollapsedCancelled && (
          <div
            ref={cancelledRef}
            className={`min-h-[60px] rounded-lg p-2 space-y-2 transition-colors ${
              isOverCancelled ? 'bg-red-50 ring-2 ring-red-300' : 'bg-gray-100'
            }`}
          >
            {cancelledIssues.map((issue) => (
              <IssueCard key={issue.uuid} issue={issue} />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Swim lane header
// ---------------------------------------------------------------------------
function SwimLaneHeader({
  label,
  count,
  isCollapsed,
  onToggle,
}: {
  label: string
  count: number
  isCollapsed: boolean
  onToggle: () => void
}) {
  return (
    <div className="col-span-full">
      <button
        onClick={onToggle}
        className="flex items-center gap-2 text-sm font-semibold text-gray-600 hover:text-gray-900 mb-2 w-full border-b border-gray-200 pb-1"
      >
        <span className="text-xs">{isCollapsed ? '▶' : '▼'}</span>
        <span className="truncate">{label}</span>
        <span className="ml-1 text-xs font-normal text-gray-400">({count})</span>
      </button>
    </div>
  )
}

// ---------------------------------------------------------------------------
// KanbanBoard — main component
// ---------------------------------------------------------------------------
interface KanbanBoardProps {
  companyUuid: string
}

export default function KanbanBoard({ companyUuid }: KanbanBoardProps) {
  const { t } = useTranslation()
  const { ws, status } = useWebSocket()
  const { data: issues = [], isLoading } = useRealtimeIssues(companyUuid)

  const [collapsed, setCollapsed] = useState<Set<string>>(loadCollapsed)
  const [draggingId, setDraggingId] = useState<string | null>(null)
  const [draggingFromStatus, setDraggingFromStatus] = useState<IssueStatus | null>(null)

  // Board settings state
  const [boardSettings, setBoardSettings] = useState<BoardSettingsData | null>(null)
  const [settingsOpen, setSettingsOpen] = useState(false)
  // Collapsed state for swim lanes (keyed by lane label)
  const [collapsedLanes, setCollapsedLanes] = useState<Set<string>>(new Set())

  // Fetch board settings on mount
  useEffect(() => {
    fetchBoardSettings(companyUuid)
      .then(setBoardSettings)
      .catch(() => {
        // Use defaults if fetch fails (e.g., migration not yet run)
        setBoardSettings({
          uuid: '',
          companyUuid,
          columnOrder: DEFAULT_COLUMN_ORDER,
          hiddenColumns: [],
          swimLaneField: null,
          createdAt: '',
          updatedAt: '',
        })
      })
  }, [companyUuid])

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } })
  )

  const toggleCollapse = useCallback((columnId: string) => {
    setCollapsed((prev) => {
      const next = new Set(prev)
      if (next.has(columnId)) next.delete(columnId)
      else next.add(columnId)
      saveCollapsed(next)
      return next
    })
  }, [])

  const toggleLane = useCallback((laneKey: string) => {
    setCollapsedLanes((prev) => {
      const next = new Set(prev)
      if (next.has(laneKey)) next.delete(laneKey)
      else next.add(laneKey)
      return next
    })
  }, [])

  const draggingIssue = draggingId ? issues.find((i) => i.uuid === draggingId) : null

  function handleDragStart(event: DragStartEvent) {
    const issue = issues.find((i) => i.uuid === event.active.id)
    if (issue) {
      setDraggingId(issue.uuid)
      setDraggingFromStatus(issue.status as IssueStatus)
    }
  }

  function handleDragOver(_event: DragOverEvent) {
    // Visual feedback handled by useDroppable isOver
  }

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event
    setDraggingId(null)
    setDraggingFromStatus(null)

    if (!over || !draggingFromStatus) return
    const toStatus = over.id as IssueStatus
    if (!isValidDrop(draggingFromStatus, toStatus)) return

    // Optimistic update happens via ws event kanban.issue.moved handled in useRealtimeIssues
    ws.moveIssue(active.id as string, toStatus)
  }

  // Compute which drop targets should be dimmed during a drag
  const dimmedColumns = useCallback(
    (columnId: string): boolean => {
      if (!draggingFromStatus) return false
      return !isValidDrop(draggingFromStatus, columnId as IssueStatus)
    },
    [draggingFromStatus]
  )

  async function handleSaveSettings(
    patch: Partial<Pick<BoardSettingsData, 'columnOrder' | 'hiddenColumns' | 'swimLaneField'>>
  ) {
    const updated = await patchBoardSettings(companyUuid, patch)
    setBoardSettings(updated)
  }

  // Derive active columns from settings
  const effectiveColumnOrder: IssueStatus[] =
    boardSettings && boardSettings.columnOrder.length > 0
      ? boardSettings.columnOrder
      : DEFAULT_COLUMN_ORDER
  const hiddenColumns = new Set(boardSettings?.hiddenColumns ?? [])
  const swimLaneField = boardSettings?.swimLaneField ?? null

  const visibleColumns = effectiveColumnOrder.filter(
    (c) => c !== 'done' && !hiddenColumns.has(c)
  )

  // Swim lane grouping: unique values of swimLaneField across issues
  const swimLaneValues: string[] = swimLaneField
    ? [...new Set(issues.map((i) => (i[swimLaneField] as string) ?? '__unassigned__'))]
    : []

  if (isLoading || boardSettings === null) {
    return <p className="text-gray-500 text-sm">{t('common.loading', 'Loading…')}</p>
  }

  function renderColumns(filteredIssues: Issue[]) {
    const getFiltered = (s: IssueStatus) => filteredIssues.filter((i) => i.status === s)
    return (
      <>
        {visibleColumns.map((col) => (
          <KanbanColumn
            key={col}
            columnId={col}
            label={t(COLUMN_LABEL_KEYS[col], col)}
            issues={getFiltered(col)}
            isDimmed={!!draggingId && dimmedColumns(col)}
            isCollapsed={collapsed.has(col)}
            onToggleCollapse={() => toggleCollapse(col)}
          />
        ))}

        {/* Done + Cancelled merged column — always shown unless hidden */}
        {!hiddenColumns.has('done') && (
          <DoneColumn
            doneIssues={getFiltered('done')}
            cancelledIssues={getFiltered('cancelled')}
            isDimmedDone={!!draggingId && dimmedColumns('done')}
            isDimmedCancelled={!!draggingId && dimmedColumns('cancelled')}
            isCollapsedDone={collapsed.has('done')}
            isCollapsedCancelled={collapsed.has('cancelled')}
            onToggleDone={() => toggleCollapse('done')}
            onToggleCancelled={() => toggleCollapse('cancelled')}
          />
        )}
      </>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-gray-800">{t('nav.issues')}</h2>
        <div className="flex items-center gap-2">
          <ConnectionBadge status={status} />
          <button
            onClick={() => setSettingsOpen(true)}
            aria-label={t('kanban.boardSettings', 'Board Settings')}
            className="rounded-md border border-gray-200 px-2 py-1 text-xs text-gray-600 hover:bg-gray-50"
          >
            {t('kanban.settings', 'Settings')}
          </button>
        </div>
      </div>

      <DndContext
        sensors={sensors}
        onDragStart={handleDragStart}
        onDragOver={handleDragOver}
        onDragEnd={handleDragEnd}
      >
        {swimLaneField ? (
          // Swim lane mode: render one row of columns per lane value
          <div className="space-y-6">
            {swimLaneValues.map((laneValue) => {
              const laneIssues = issues.filter(
                (i) => ((i[swimLaneField] as string) ?? '__unassigned__') === laneValue
              )
              const label =
                laneValue === '__unassigned__'
                  ? t('kanban.unassigned', 'Unassigned')
                  : laneValue
              return (
                <div key={laneValue}>
                  <SwimLaneHeader
                    label={label}
                    count={laneIssues.length}
                    isCollapsed={collapsedLanes.has(laneValue)}
                    onToggle={() => toggleLane(laneValue)}
                  />
                  {!collapsedLanes.has(laneValue) && (
                    <div className="flex gap-4 overflow-x-auto pb-2">
                      {renderColumns(laneIssues)}
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        ) : (
          // Normal (flat) mode
          <div className="flex gap-4 overflow-x-auto pb-4">
            {renderColumns(issues)}
          </div>
        )}

        <DragOverlay>
          {draggingIssue ? <IssueCard issue={draggingIssue} isDragging /> : null}
        </DragOverlay>
      </DndContext>

      {settingsOpen && boardSettings && (
        <BoardSettingsDialog
          settings={boardSettings}
          onClose={() => setSettingsOpen(false)}
          onSave={handleSaveSettings}
        />
      )}
    </div>
  )
}
