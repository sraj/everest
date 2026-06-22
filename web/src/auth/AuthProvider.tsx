import { createContext, useContext, useEffect, useState, useCallback, type ReactNode } from 'react'
import { UserManager, User } from 'oidc-client-ts'
import { authConfig } from './authConfig'

interface AuthContextValue {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  login: () => Promise<void>
  logout: () => Promise<void>
  getAccessToken: () => string | undefined
}

const AuthContext = createContext<AuthContextValue | null>(null)

let userManager: UserManager | null = null

function getUserManager(): UserManager | null {
  if (!authConfig.client_id) return null
  if (!userManager) {
    userManager = new UserManager(authConfig)
  }
  return userManager
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const mgr = getUserManager()

  const login = useCallback(async () => {
    const m = getUserManager()
    if (m) await m.signinRedirect()
  }, [])

  const logout = useCallback(async () => {
    const m = getUserManager()
    if (m) {
      setUser(null)
      await m.signoutRedirect()
    }
  }, [])

  const getAccessToken = useCallback(() => {
    return user?.id_token
  }, [user])

  useEffect(() => {
    if (!mgr) {
      setIsLoading(false)
      return
    }

    // Handle redirect from Zitadel
    mgr.signinRedirectCallback().then((u) => {
      setUser(u)
      setIsLoading(false)
      window.history.replaceState({}, document.title, '/')
    }).catch(() => {
      // Not a redirect callback, load existing user
      mgr.getUser().then((u) => {
        setUser(u ?? null)
        setIsLoading(false)
      })
    })

    // Auto-renew on token refresh
    mgr.events.addUserLoaded((u) => setUser(u))
    mgr.events.addUserUnloaded(() => setUser(null))
  }, [mgr])

  return (
    <AuthContext.Provider value={{ user, isLoading, isAuthenticated: !!user, login, logout, getAccessToken }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
