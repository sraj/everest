import { EverstLogo } from './icons'

export function LandingHero({ onLogin }: { onLogin: () => void }) {
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <header className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center">
          <div className="flex items-center gap-3">
            <EverstLogo className="w-8 h-8 text-blue-600" />
            <h1 className="text-xl font-semibold text-gray-900 dark:text-white">Everest Docs</h1>
          </div>
        </div>
      </header>
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 text-center">
        <EverstLogo className="mx-auto h-20 w-20 text-blue-600" />
        <h2 className="mt-6 text-3xl font-bold text-gray-900 dark:text-white">Welcome to Everest Docs</h2>
        <p className="mt-4 text-lg text-gray-500 dark:text-gray-400 max-w-md mx-auto">
          Sign in to create, edit, and manage your documents.
        </p>
        <button
          onClick={onLogin}
          className="mt-8 inline-flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-8 py-3 rounded-lg text-lg font-medium transition-colors shadow-sm"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
          </svg>
          Sign in
        </button>
      </main>
    </div>
  )
}
