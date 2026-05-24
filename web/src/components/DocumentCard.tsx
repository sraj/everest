import { Link } from 'react-router-dom'
import { DeleteIcon, ClockIcon, SpinnerIcon } from './icons'
import type { Document } from '../store/documentSlice'

type DocumentCardProps = {
  doc: Document
  deletingId: string | null
  onDelete: (e: React.MouseEvent, id: string) => void
}

export function DocumentCard({ doc, deletingId, onDelete }: DocumentCardProps) {
  return (
    <Link
      key={doc.id}
      to={`/documents/${doc.id}`}
      className="group block bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden hover:border-blue-500 dark:hover:border-blue-500 hover:shadow-lg transition-all"
    >
      <div className="aspect-[3/4] bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 overflow-hidden relative">
        {doc.thumbnail_url ? (
          <img
            src={doc.thumbnail_url}
            alt={doc.title || 'Document preview'}
            className="w-full h-full object-cover object-top"
            loading="lazy"
          />
        ) : (
          <div
            className="absolute inset-0 origin-top-left scale-[0.4] w-[250%] p-6 text-sm text-gray-700 dark:text-gray-300 leading-relaxed prose prose-sm dark:prose-invert max-w-none pointer-events-none"
            dangerouslySetInnerHTML={{ __html: doc.content || '<p class="text-gray-400">Empty document</p>' }}
          />
        )}
      </div>
      <div className="p-4">
        <div className="flex items-start justify-between gap-2">
          <h3 className="font-medium text-gray-900 dark:text-white truncate group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">
            {doc.title || 'Untitled Document'}
          </h3>
          <button
            onClick={(e) => onDelete(e, doc.id)}
            disabled={deletingId === doc.id}
            className="flex-shrink-0 p-1.5 text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-md transition-colors disabled:opacity-50"
            title="Delete document"
          >
            {deletingId === doc.id ? (
              <SpinnerIcon className="w-4 h-4 text-red-500" />
            ) : (
              <DeleteIcon className="w-4 h-4" />
            )}
          </button>
        </div>
        <div className="mt-2 flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <ClockIcon className="w-4 h-4" />
          <span>
            {new Date(doc.updated_at).toLocaleDateString(undefined, {
              month: 'short',
              day: 'numeric',
              year: 'numeric',
            })}
          </span>
        </div>
      </div>
    </Link>
  )
}
