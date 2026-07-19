import { useState, useEffect, useCallback } from 'react'
import { permissions, type PermissionResponse } from '../../lib/api'
import { useAuth } from '../../contexts/AuthContext'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Modal } from '../../components/ui/Modal'

export function PermissionsPage() {
  const { hasPermission, hasRole } = useAuth()
  const [permList, setPermList] = useState<PermissionResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [newPerm, setNewPerm] = useState({ name: '', description: '' })
  const [creating, setCreating] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editData, setEditData] = useState({ name: '', description: '' })
  const [search, setSearch] = useState('')

  const canCreate = hasRole('super_admin') || hasPermission('permissions:create')
  const canUpdate = hasRole('super_admin') || hasPermission('permissions:update')
  const canDelete = hasRole('super_admin') || hasPermission('permissions:delete')

  const filteredPerms = permList.filter(p => {
    const q = search.toLowerCase()
    return p.name.toLowerCase().includes(q) || p.description.toLowerCase().includes(q)
  })

  const fetchPerms = useCallback(async () => {
    setLoading(true)
    try {
      const data = await permissions.list()
      setPermList(data)
    } catch { /* ignore */ }
    setLoading(false)
  }, [])

  useEffect(() => {
    fetchPerms()
  }, [fetchPerms])

  async function handleCreate() {
    if (!newPerm.name || !newPerm.description) return
    setCreating(true)
    try {
      await permissions.create(newPerm)
      setShowCreate(false)
      setNewPerm({ name: '', description: '' })
      fetchPerms()
    } catch { /* ignore */ }
    setCreating(false)
  }

  async function handleUpdate(id: number) {
    try {
      await permissions.update(id, editData)
      setEditingId(null)
      fetchPerms()
    } catch { /* ignore */ }
  }

  async function handleDelete(id: number, name: string) {
    if (!confirm(`Delete permission "${name}"?`)) return
    try {
      await permissions.delete(id)
      fetchPerms()
    } catch { /* ignore */ }
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <Input
          placeholder="Search permissions..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-gray-900">Permissions ({filteredPerms.length})</h3>
          {canCreate && (
            <Button size="sm" onClick={() => setShowCreate(true)}>
              Create Permission
            </Button>
          )}
        </div>
        {loading ? (
          <div className="p-6 text-center text-gray-500">Loading...</div>
        ) : filteredPerms.length === 0 ? (
          <div className="p-6 text-center text-gray-500">No permissions found</div>
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
                {filteredPerms.map(p => (
                  <tr key={p.id} className="hover:bg-gray-50">
                    <td className="px-6 py-3">
                      {editingId === p.id ? (
                        <Input
                          value={editData.name}
                          onChange={(e) => setEditData(prev => ({ ...prev, name: e.target.value }))}
                        />
                      ) : (
                        <span className="font-medium text-gray-900">{p.name}</span>
                      )}
                    </td>
                    <td className="px-6 py-3">
                      {editingId === p.id ? (
                        <Input
                          value={editData.description}
                          onChange={(e) => setEditData(prev => ({ ...prev, description: e.target.value }))}
                        />
                      ) : (
                        <span className="text-gray-600">{p.description}</span>
                      )}
                    </td>
                    <td className="px-6 py-3 text-gray-500">{p.createdBy}</td>
                    <td className="px-6 py-3">
                      <div className="flex gap-2">
                        {canUpdate && (
                          editingId === p.id ? (
                            <>
                              <Button variant="ghost" size="sm" onClick={() => handleUpdate(p.id)}>Save</Button>
                              <Button variant="ghost" size="sm" onClick={() => setEditingId(null)}>Cancel</Button>
                            </>
                          ) : (
                            <Button variant="ghost" size="sm" onClick={() => {
                              setEditingId(p.id)
                              setEditData({ name: p.name, description: p.description })
                            }}>
                              Edit
                            </Button>
                          )
                        )}
                        {canDelete && (
                          <Button variant="danger" size="sm" onClick={() => handleDelete(p.id, p.name)}>
                            Delete
                          </Button>
                        )}
                      </div>
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
        title="Create Permission"
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
            value={newPerm.name}
            onChange={(e) => setNewPerm(prev => ({ ...prev, name: e.target.value }))}
            placeholder="e.g. posts:read"
          />
          <Input
            label="Description"
            value={newPerm.description}
            onChange={(e) => setNewPerm(prev => ({ ...prev, description: e.target.value }))}
            placeholder="Permission description"
          />
        </div>
      </Modal>
    </div>
  )
}
