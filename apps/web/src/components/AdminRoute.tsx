import { Navigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

interface AdminRouteProps {
  children: React.ReactNode
  requirePermission?: string
  requireAnyPermission?: string[]
}

export function AdminRoute({ children, requirePermission, requireAnyPermission }: AdminRouteProps) {
  const { hasPermission, hasRole, hasAnyPermission, isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-500">Loading...</div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  if (hasRole('super_admin')) {
    return <>{children}</>
  }

  if (requirePermission && !hasPermission(requirePermission)) {
    return <Navigate to="/" replace />
  }

  if (requireAnyPermission && !hasAnyPermission(...requireAnyPermission)) {
    return <Navigate to="/" replace />
  }

  return <>{children}</>
}
