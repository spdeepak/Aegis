import { useAuth } from '../contexts/AuthContext'
import { Badge } from '../components/ui/Badge'

export function Dashboard() {
  const { user } = useAuth()

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-2">
          Welcome, {user?.firstName} {user?.lastName}
        </h2>
        <p className="text-gray-500 text-sm">
          You are signed in as <span className="font-medium text-gray-700">{user?.email}</span>
        </p>

        <div className="mt-4">
          <h3 className="text-sm font-medium text-gray-700 mb-2">Your Roles</h3>
          <div className="flex flex-wrap gap-2">
            {user?.roles.length ? (
              user.roles.map(role => (
                <Badge key={role} variant="info">{role}</Badge>
              ))
            ) : (
              <span className="text-sm text-gray-400">No roles assigned</span>
            )}
          </div>
        </div>

        <div className="mt-4">
          <h3 className="text-sm font-medium text-gray-700 mb-2">Your Permissions</h3>
          <div className="flex flex-wrap gap-2">
            {user?.permissions.length ? (
              user.permissions.map(perm => (
                <Badge key={perm}>{perm}</Badge>
              ))
            ) : (
              <span className="text-sm text-gray-400">No permissions assigned</span>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
