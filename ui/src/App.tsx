import { useState } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import LanguageSwitcher from './components/LanguageSwitcher'
import { WebSocketProvider } from './context/WebSocketProvider'
import KanbanBoard from './components/KanbanBoard'
import SettingsPanel from './components/SettingsPanel'

const queryClient = new QueryClient()

// Default company UUID — in production this would come from auth/routing context
const COMPANY_UUID = import.meta.env.VITE_COMPANY_UUID || 'default'

type Page = 'kanban' | 'settings'

function AppContent() {
  const { t } = useTranslation()
  const [page, setPage] = useState<Page>('kanban')

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between">
        <h1 className="text-xl font-semibold text-gray-900">{t('app.title')}</h1>
        <div className="flex items-center gap-4">
          <LanguageSwitcher />
          <button
            onClick={() => setPage(page === 'settings' ? 'kanban' : 'settings')}
            className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
            aria-current={page === 'settings' ? 'page' : undefined}
          >
            {t('nav.settings')}
          </button>
        </div>
      </header>
      <main className="p-6">
        {page === 'settings' ? (
          <SettingsPanel />
        ) : (
          <KanbanBoard companyUuid={COMPANY_UUID} />
        )}
      </main>
    </div>
  )
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <WebSocketProvider companyUuid={COMPANY_UUID}>
        <AppContent />
      </WebSocketProvider>
    </QueryClientProvider>
  )
}
