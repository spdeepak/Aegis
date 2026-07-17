import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { roles, type RoleResponse } from '../../lib/api'
import { useAuth } from '../../contexts/AuthContext'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Modal } from '../../components/ui/Modal'

export function RolesPage() {
  const { hasPermission, hasRole } = useAuth()
  const [roleList, setRoleList] = useState<RoleResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [newRole, setNewRole] = useState({ name: '', description: '' })
  const [creating, setCreating] = useState(false)

  const canCreate = hasRole('super_admin') || hasPermission('roles:create')
  const canDelete = hasRole('super_admin') || hasPermission('roles:delete')

  const fetchRoles = useCallback(async () => {
    setLoading(true)
    try {
      const data = await roles.list()
      setRoleList(data)
    } catch { /* ignore */ }
    setLoading(false)
  }, [])

  useEffect(() => {
    fetchRoles()
  }, [fetchRoles])

  async function handleCreate() {
    if (!newRole.name || !newRole.description) return
    setCreating(true)
    try {
      await roles.create(newRole)
      setShowCreate(false)
      setNewRole({ name: '', description: '' })
      fetchRoles()
    } catch { /* ignore */ }
    setCreating(false)
  }

  async function handleDelete(id: number, name: string) {
    if (!confirm(`Delete role "${name}"?`)) return
    try {
      await roles.delete(id)
      fetchRoles()
    } catch { /* ignore */ }
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-xl shadow-sm border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-gray-900">Roles ({roleList.length})</h3>
          {canCreate && (
            <Button size="sm" onClick={() => setShowCreate(true)}>
              Create Role
            </Button>
          )}
        </div>
        {loading ? (
          <div className="p-6 text-center text-gray-500">Loading...</div>
        ) : roleList.length === 0 ? (
          <div className="p-6 text-center text-gray-500">No roles found</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 text-left">
                <tr>
                  <th className="px-6 py-3 font-medium text-gray-500">Name</th>
                  <th className="px-6 py-3 font-medium text-gray-500">Description</th>
                  <th className="px-6 py-3 font-medium text-gray-500">Created By</th>
                  <th className="px-6 py-3 font-medium text-gray-500">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {roleList.map(r => (
                  <tr key={r.id} className="hover:bg-gray-50">
                    <td className="px-6 py-3">
                      <Link to={`/admin/roles/${r.id}`} className="font-medium text-gray-900 hover:text-indigo-600">
                        {r.name}
                      </Link>
                    </td>
                    <td className="px-6 py-3 text-gray-600">{r.description}</td>
                    <td className="px-6 py-3 text-gray-500">{r.createdBy}</td>
                    <td className="px-6 py-3">
                      {canDelete && (
                        <Button variant="danger" size="sm" onClick={() => handleDelete(r.id, r.name)}>
                          Delete
                        </Button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <Modal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        title="Create Role"
        footer={
          <>
            <Button variant="secondary" onClick={() => setShowCreate(false)}>Cancel</Button>
            <Button onClick={handleCreate} disabled={creating}>
              {creating ? 'Creating...' : 'Create'}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            label="Name"
            value={newRole.name}
            onChange={(e) => setNewRole(prev => ({ ...prev, name: e.target.value }))}
            placeholder="e.g. editor"
          />
          <Input
            label="Description"
            value={newRole.description}
            onChange={(e) => setNewRole(prev => ({ ...prev, description: e.target.value }))}
            placeholder="Role description"
          />
        </div>
      </Modal>
    </div>
  )
}
