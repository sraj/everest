import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Editor } from '../components/editor'

export function DocumentEditor() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isNew = id === 'new'

  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [saving, setSaving] = useState(false)
  const [loading, setLoading] = useState(!isNew)
  const [lastSaved, setLastSaved] = useState<Date | null>(null)

  // Load existing document
  useEffect(() => {
    if (!isNew && id) {
      setLoading(true)
      fetch(`/api/v1/documents/${id}`)
        .then((res) => res.json())
        .then((data) => {
          setTitle(data.title || '')
          setContent(data.content || '')
        })
        .catch((err) => console.error('Failed to load document:', err))
        .finally(() => setLoading(false))
    }
  }, [id, isNew])

  const handleSave = async () => {
    setSaving(true)
    try {
      const url = isNew ? '/api/v1/documents' : `/api/v1/documents/${id}`
      const method = isNew ? 'POST' : 'PUT'

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ title: title || 'Untitled Document', content }),
      })

      if (response.ok) {
        setLastSaved(new Date())
        if (isNew) {
          const data = await response.json()
          navigate(`/documents/${data.id}`, { replace: true })
        }
      }
    } catch (error) {
      console.error('Failed to save document:', error)
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-500">Loading document...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="h-screen flex flex-col bg-gray-100 dark:bg-gray-900">
      {/* Header */}
      <header className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/')}
            className="p-2 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
            title="Back to Documents"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
          </button>

          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Untitled Document"
            className="text-lg font-medium bg-transparent border-none focus:outline-none focus:ring-0 text-gray-900 dark:text-gray-100 placeholder-gray-400 w-64 md:w-96"
          />
        </div>

        <div className="flex items-center gap-4">
          {lastSaved && (
            <span className="text-sm text-gray-500 hidden md:block">
              Last saved: {lastSaved.toLocaleTimeString()}
            </span>
          )}

          <button
            onClick={handleSave}
            disabled={saving}
            className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white px-4 py-2 rounded-lg transition-colors"
          >
            {saving ? (
              <>
                <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                </svg>
                Saving...
              </>
            ) : (
              <>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
                Save
              </>
            )}
          </button>
        </div>
      </header>

      {/* Editor */}
      <div className="flex-1 overflow-hidden">
        <Editor key={id} content={content} onChange={setContent} />
      </div>
    </div>
  )
}
