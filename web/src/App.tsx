import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Provider } from 'react-redux'

import { store } from './store'
import { Home } from './pages/Home'
import { DocumentEditor } from './pages/DocumentEditor'

function App() {
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
