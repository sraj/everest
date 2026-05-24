import { useEffect, useState } from 'react'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { fetchDocuments, deleteDocument } from '../store/documentSlice'
import { useAuth } from '../auth/AuthProvider'
import { LandingHero } from '../components/LandingHero'
import { AppHeader } from '../components/AppHeader'
import { DocumentGrid } from '../components/DocumentGrid'

export function Home() {
  const dispatch = useAppDispatch()
  const { documents, loading, error } = useAppSelector((state) => state.documents)
  const [deletingId, setDeletingId] = useState<string | null>(null)
  const { user, isAuthenticated, isLoading, login, logout, getAccessToken } = useAuth()

  useEffect(() => {
    if (isAuthenticated) {
      dispatch(fetchDocuments(getAccessToken()))
    }
  }, [dispatch, isAuthenticated, getAccessToken])

  const handleDelete = async (e: React.MouseEvent, id: string) => {
    e.preventDefault()
    e.stopPropagation()

    if (confirm('Are you sure you want to delete this document?')) {
      setDeletingId(id)
      try {
        await dispatch(deleteDocument({ id, accessToken: getAccessToken() })).unwrap()
      } finally {
        setDeletingId(null)
      }
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
      </div>
    )
  }

  if (!isAuthenticated) {
    return <LandingHero onLogin={login} />
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <AppHeader user={user} onLogout={logout} />
      <DocumentGrid
        documents={documents}
        loading={loading}
        error={error}
        deletingId={deletingId}
        onDelete={handleDelete}
      />
    </div>
  )
}
