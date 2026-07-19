import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react'
import { getUserFromToken, isTokenExpired, type DecodedUser } from '../lib/jwt'
import { getAccessToken, getRefreshToken, setTokens, clearTokens, refreshAccessToken } from '../lib/api'

interface AuthContextType {
  user: DecodedUser | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (accessToken: string, refreshToken: string) => void
  logout: () => void
  hasPermission: (permission: string) => boolean
  hasRole: (role: string) => boolean
  hasAnyPermission: (...permissions: string[]) => boolean
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<DecodedUser | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  const loadUser = useCallback(async () => {
    const token = getAccessToken()
    if (!token) {
      setUser(null)
      setIsLoading(false)
      return
    }

    if (isTokenExpired(token)) {
      const refreshToken = getRefreshToken()
      if (refreshToken && !isTokenExpired(refreshToken)) {
        try {
          const newToken = await refreshAccessToken()
          setUser(getUserFromToken(newToken))
        } catch {
          clearTokens()
          setUser(null)
        }
      } else {
        clearTokens()
        setUser(null)
      }
    } else {
      setUser(getUserFromToken(token))
    }
    setIsLoading(false)
  }, [])

  useEffect(() => {
    loadUser()
  }, [loadUser])

  const login = useCallback((accessToken: string, refreshToken: string) => {
    setTokens(accessToken, refreshToken)
    setUser(getUserFromToken(accessToken))
  }, [])

  const logout = useCallback(() => {
    clearTokens()
    setUser(null)
  }, [])

  const hasPermission = useCallback((permission: string) => {
    return user?.permissions.includes(permission) ?? false
  }, [user])

  const hasRole = useCallback((role: string) => {
    return user?.roles.includes(role) ?? false
  }, [user])

  const hasAnyPermission = useCallback((...permissions: string[]) => {
    return permissions.some(p => user?.permissions.includes(p) ?? false)
  }, [user])

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        logout,
        hasPermission,
        hasRole,
        hasAnyPermission,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
