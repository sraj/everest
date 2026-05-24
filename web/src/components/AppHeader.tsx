import { Link } from 'react-router-dom'
import { EverstLogo, PlusIcon } from './icons'
import type { User } from 'oidc-client-ts'

type AppHeaderProps = {
  user: User | null
  onLogout: () => void
}

export function AppHeader({ user, onLogout }: AppHeaderProps) {
  return (
    <header className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 sticky top-0 z-10">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center gap-3">
            <EverstLogo className="w-8 h-8 text-blue-600" />
            <h1 className="text-xl font-semibold text-gray-900 dark:text-white">Everest Docs</h1>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600 dark:text-gray-300 hidden sm:block">
              {user?.profile?.name || user?.profile?.email || 'User'}
            </span>
            <Link
              to="/documents/new"
              className="inline-flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg transition-colors shadow-sm"
            >
              <PlusIcon className="w-5 h-5" />
              New Document
            </Link>
            <button
              onClick={onLogout}
              className="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-white transition-colors"
            >
              Logout
            </button>
          </div>
        </div>
      </div>
    </header>
  )
}
