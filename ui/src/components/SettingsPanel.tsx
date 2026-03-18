import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import LanguageSwitcher from './LanguageSwitcher'

type CommLang = 'en' | 'ko' | 'ja' | null

const COMM_LANG_OPTIONS: { value: CommLang; label: string }[] = [
  { value: 'en', label: 'English' },
  { value: 'ko', label: '한국어' },
  { value: 'ja', label: '日本語' },
  { value: null, label: '' }, // label set via t()
]

export default function SettingsPanel() {
  const { t } = useTranslation('settings')
  const [commLang, setCommLang] = useState<CommLang>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/instance-settings')
      .then((res) => res.json())
      .then((data) => {
        setCommLang(data.communicationLanguage ?? null)
      })
      .catch(() => {
        // silently fall back to null (follow UI)
      })
      .finally(() => setLoading(false))
  }, [])

  function handleCommLangChange(value: string) {
    const next: CommLang = value === '' ? null : (value as 'en' | 'ko' | 'ja')
    setCommLang(next)
    fetch('/api/instance-settings', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ communicationLanguage: next }),
    }).catch(() => {
      // best-effort persist; revert optimistic update on error
      setCommLang(commLang)
    })
  }

  return (
    <div className="max-w-2xl mx-auto p-6 space-y-8">
      {/* Section 1: UI Language */}
      <section>
        <h2 className="text-base font-semibold text-gray-900 mb-1">
          {t('uiLanguage')}
        </h2>
        <p className="text-sm text-gray-500 mb-3">{t('uiLanguageDescription')}</p>
        <LanguageSwitcher />
      </section>

      <hr className="border-gray-200" />

      {/* Section 2: Communication Language */}
      <section>
        <h2 className="text-base font-semibold text-gray-900 mb-1">
          {t('communicationLanguage')}
        </h2>
        <p className="text-sm text-gray-500 mb-3">
          {t('communicationLanguageDescription')}
        </p>
        <select
          value={commLang ?? ''}
          onChange={(e) => handleCommLangChange(e.target.value)}
          disabled={loading}
          className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
        >
          {COMM_LANG_OPTIONS.map((opt) => (
            <option key={String(opt.value)} value={opt.value ?? ''}>
              {opt.value === null ? t('followUiLanguage') : opt.label}
            </option>
          ))}
        </select>
      </section>
    </div>
  )
}
