import { useEffect, useRef } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Provider } from 'react-redux'
import { init, type WidgetInstance } from '@srajvenkat/aichat-widget'

import { store } from './store'
import { Home } from './pages/Home'
import { DocumentEditor } from './pages/DocumentEditor'

function App() {
  const widgetRef = useRef<WidgetInstance | null>(null)

  useEffect(() => {
    if (widgetRef.current) return

    widgetRef.current = init({
      apiKey: import.meta.env.VITE_AI_CHAT_API_KEY || '',
      connection: {
        protocol: 'sse',
        baseUrl: window.location.origin,
        sseEndpoint: '/api/chat/stream',
        messagesEndpoint: '/api/chat/messages',
      },
      position: {
        position: 'bottom-right',
        offsetBottom: '80px',
      },
      theme: {
        darkMode: true,
      },
      title: 'Everest AI Assistant',
    })

    return () => {
      widgetRef.current?.destroy()
      widgetRef.current = null
    }
  }, [])

  return (
    <Provider store={store}>
      <BrowserRouter>
        <div className="min-h-screen bg-gray-100 dark:bg-gray-900">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/documents/:id" element={<DocumentEditor />} />
          </Routes>
        </div>
      </BrowserRouter>
    </Provider>
  )
}

export default App
