import { Suspense } from 'react'
import { createRoot } from 'react-dom/client'
import './i18n/config'
import './index.css'
import App from './App'

createRoot(document.getElementById('root')!).render(
  <Suspense fallback={<div>Loading...</div>}>
    <App />
  </Suspense>,
)
