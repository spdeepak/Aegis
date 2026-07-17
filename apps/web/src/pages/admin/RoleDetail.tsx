import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { roles, permissions, type RoleResponse, type PermissionResponse } from '../../lib/api'
import { useAuth } from '../../contexts/AuthContext'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'

export function RoleDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { hasPermission, hasRole } = useAuth()
  const roleId = Number(id)

  const [role, setRole] = useState<RoleResponse | null>(null)
  const [rolePerms, setRolePerms] = useState<PermissionResponse[]>([])
  const [allPerms, setAllPerms] = useState<PermissionResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedPermIds, setSelectedPermIds] = useState<number[]>([])
  const [editing, setEditing] = useState(false)
  const [editData, setEditData] = useState({ name: '', description: '' })

  const canUpdate = hasRole('super_admin') || hasPermission('roles:update')
  const canAssignPerms = hasRole('super_admin') || hasPermission('roles:assign-permission')
  const canUnassignPerms = hasRole('super_admin') || hasPermission('roles:unassign-permission')

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [r, rp, ap] = await Promise.all([
        roles.get(roleId),
        roles.listAndPermissions(),
        permissions.list(),
      ])
      setRole(r)
      setEditData({ name: r.name, description: r.description })
      const roleEntry = rp.find((entry: { roles: { id: number } }) => entry.roles.id === roleId)
      setRolePerms(roleEntry?.roles.permissions || [])
      setAllPerms(ap)
    } catch {
      navigate('/admin/roles')
    }
    setLoading(false)
  }, [roleId, navigate])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  async function handleUpdate() {
    try {
      await roles.update(roleId, editData)
      setEditing(false)
      fetchData()
    } catch { /* ignore */ }
  }

  async function handleAssignPermissions() {
    if (selectedPermIds.length === 0) return
    try {
      await roles.assignPermissions(roleId, selectedPermIds)
      setSelectedPermIds([])
      fetchData()
    } catch { /* ignore */ }
  }

  async function handleUnassignPermission(permId: number) {
    try {
      await roles.unassignPermission(roleId, permId)
      fetchData()
    } catch { /* ignore */ }
  }

  if (loading) {
    return <div className="text-center text-gray-500 py-8">Loading...</div>
  }

  if (!role) return null

  const assignedPermIds = new Set(rolePerms.map(p => p.id))
  const unassignedPerms = allPerms.filter(p => !assignedPermIds.has(p.id))

  return (
    <div className="space-y-6 max-w-2xl">
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-4">
          {editing ? (
            <div className="flex-1 space-y-3">
              <Input
                label="Name"
                value={editData.name}
                onChange={(e) => setEditData(prev => ({ ...prev, name: e.target.value }))}
              />
              <Input
                label="Description"
                value={editData.description}
                onChange={(e) => setEditData(prev => ({ ...prev, description: e.target.value }))}
              />
              <div className="flex gap-2">
                <Button size="sm" onClick={handleUpdate}>Save</Button>
                <Button variant="secondary" size="sm" onClick={() => setEditing(false)}>Cancel</Button>
              </div>
            </div>
          ) : (
            <>
              <div>
                <h2 className="text-lg font-semibold text-gray-900">{role.name}</h2>
                <p className="text-sm text-gray-500">{role.description}</p>
              </div>
              {canUpdate && (
                <Button variant="secondary" size="sm" onClick={() => setEditing(true)}>
                  Edit
                </Button>
              )}
            </>
          )}
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Permissions ({rolePerms.length})</h3>
        <div className="flex flex-wrap gap-2 mb-4">
          {rolePerms.length > 0 ? (
            rolePerms.map(p => (
              <div key={p.id} className="flex items-center gap-1">
                <Badge>{p.name}</Badge>
                {canUnassignPerms && (
                  <button
                    onClick={() => handleUnassignPermission(p.id)}
                    className="text-gray-400 hover:text-red-500 text-xs"
                  >
                    &times;
                  </button>
                )}
              </div>
            ))
          ) : (
            <span className="text-sm text-gray-400">No permissions assigned</span>
          )}
        </div>

        {canAssignPerms && unassignedPerms.length > 0 && (
          <div className="flex gap-2 items-end">
            <select
              className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
              value={selectedPermIds[0] || ''}
              onChange={(e) => setSelectedPermIds(e.target.value ? [Number(e.target.value)] : [])}
            >
              <option value="">Select a permission</option>
              {unassignedPerms.map(p => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
            <Button size="sm" onClick={handleAssignPermissions} disabled={selectedPermIds.length === 0}>
              Assign
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
