import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { users, roles, type UserWithRoles, type RoleResponse } from '../../lib/api'
import { useAuth } from '../../contexts/AuthContext'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'

export function UserDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { hasPermission, hasRole } = useAuth()
  const userId = Number(id)

  const [userRoles, setUserRoles] = useState<UserWithRoles | null>(null)
  const [allRoles, setAllRoles] = useState<RoleResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedRoleIds, setSelectedRoleIds] = useState<number[]>([])

  const canManage = hasRole('super_admin') || hasPermission('users:assign-roles')

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [ur, rl] = await Promise.all([users.getRoles(userId), roles.list()])
      setUserRoles(ur)
      setAllRoles(rl)
    } catch {
      navigate('/admin/users')
    }
    setLoading(false)
  }, [userId, navigate])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  async function handleAssignRoles() {
    if (selectedRoleIds.length === 0) return
    try {
      await users.assignRoles(userId, selectedRoleIds)
      setSelectedRoleIds([])
      fetchData()
    } catch { /* ignore */ }
  }

  async function handleRemoveRole(roleId: number) {
    try {
      await users.removeRole(userId, roleId)
      fetchData()
    } catch { /* ignore */ }
  }

  async function handleToggleLock() {
    if (!userRoles) return
    try {
      await users.lock(userId)
      fetchData()
    } catch { /* ignore */ }
  }

  async function handleUnlock() {
    try {
      await users.unlock(userId)
      fetchData()
    } catch { /* ignore */ }
  }

  if (loading) {
    return <div className="text-center text-gray-500 py-8">Loading...</div>
  }

  if (!userRoles) return null

  const unassignedRoles = allRoles.filter(r => !userRoles.roles.includes(r.name))

  return (
    <div className="space-y-6 max-w-2xl">
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900">
            {userRoles.firstName} {userRoles.lastName}
          </h2>
          <div className="flex gap-2">
            {canManage && (
              <>
                <Button variant="secondary" size="sm" onClick={handleToggleLock}>Lock</Button>
                <Button variant="secondary" size="sm" onClick={handleUnlock}>Unlock</Button>
              </>
            )}
          </div>
        </div>
        <dl className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <dt className="text-gray-500">Email</dt>
            <dd className="font-medium text-gray-900">{userRoles.email}</dd>
          </div>
          <div>
            <dt className="text-gray-500">User ID</dt>
            <dd className="font-medium text-gray-900">{userRoles.id}</dd>
          </div>
        </dl>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Roles</h3>
        <div className="flex flex-wrap gap-2 mb-4">
          {userRoles.roles.length > 0 ? (
            userRoles.roles.map(r => (
              <div key={r} className="flex items-center gap-1">
                <Badge variant="info">{r}</Badge>
                {canManage && (
                  <button
                    onClick={() => {
                      const role = allRoles.find(ro => ro.name === r)
                      if (role) handleRemoveRole(role.id)
                    }}
                    className="text-gray-400 hover:text-red-500 text-xs"
                  >
                    &times;
                  </button>
                )}
              </div>
            ))
          ) : (
            <span className="text-sm text-gray-400">No roles assigned</span>
          )}
        </div>

        {canManage && unassignedRoles.length > 0 && (
          <div className="flex gap-2 items-end">
            <select
              className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
              value={selectedRoleIds[0] || ''}
              onChange={(e) => setSelectedRoleIds(e.target.value ? [Number(e.target.value)] : [])}
            >
              <option value="">Select a role to assign</option>
              {unassignedRoles.map(r => (
                <option key={r.id} value={r.id}>{r.name}</option>
              ))}
            </select>
            <Button size="sm" onClick={handleAssignRoles} disabled={selectedRoleIds.length === 0}>
              Assign
            </Button>
          </div>
        )}
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Permissions</h3>
        <div className="flex flex-wrap gap-2">
          {userRoles.permissions.length > 0 ? (
            userRoles.permissions.map(p => (
              <Badge key={p}>{p}</Badge>
            ))
          ) : (
            <span className="text-sm text-gray-400">No direct permissions assigned</span>
          )}
        </div>
      </div>
    </div>
  )
}
