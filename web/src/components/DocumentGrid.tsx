import { Link } from 'react-router-dom'
import { DocumentIcon, PlusIcon } from './icons'
import { DocumentCard } from './DocumentCard'
import type { Document } from '../store/documentSlice'

type DocumentGridProps = {
  documents: Document[]
  loading: boolean
  error: string | null
  deletingId: string | null
  onDelete: (e: React.MouseEvent, id: string) => void
}

export function DocumentGrid({ documents, loading, error, deletingId, onDelete }: DocumentGridProps) {
  return (
    <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-6">
        <h2 className="text-lg font-medium text-gray-900 dark:text-white mb-4">Recent Documents</h2>
      </div>

      {loading && (
        <div className="flex justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
        </div>
      )}

      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 p-4 rounded-lg">
          {error}
        </div>
      )}

      {!loading && documents.length === 0 && (
        <div className="text-center py-16">
          <DocumentIcon className="mx-auto h-16 w-16 text-gray-300 dark:text-gray-600" />
          <h3 className="mt-4 text-lg font-medium text-gray-900 dark:text-white">No documents yet</h3>
          <p className="mt-2 text-gray-500 dark:text-gray-400">Get started by creating your first document.</p>
          <Link
            to="/documents/new"
            className="mt-6 inline-flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-6 py-3 rounded-lg transition-colors"
          >
            <PlusIcon className="w-5 h-5" />
            Create Document
          </Link>
        </div>
      )}

      <div className="grid gap-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6">
        {documents.map((doc) => (
          <DocumentCard key={doc.id} doc={doc} deletingId={deletingId} onDelete={onDelete} />
        ))}
      </div>
    </main>
  )
}
