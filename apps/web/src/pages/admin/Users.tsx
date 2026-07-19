import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { users, type UserDetails } from '../../lib/api'
import { useAuth } from '../../contexts/AuthContext'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'

export function UsersPage() {
  const { user: currentUser } = useAuth()
  const [userList, setUserList] = useState<UserDetails[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState({ firstName: '', lastName: '', email: '' })

  const fetchUsers = useCallback(async () => {
    setLoading(true)
    try {
      const data = await users.list({
        ...search,
        size: 50,
        page: 0,
      })
      setUserList(data)
    } catch { /* ignore */ }
    setLoading(false)
  }, [search])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  function handleSearch(field: string, value: string) {
    setSearch(prev => ({ ...prev, [field]: value }))
  }

  async function handleToggleLock(u: UserDetails) {
    try {
      if (u.locked) {
        await users.unlock(u.id)
      } else {
        await users.lock(u.id)
      }
      fetchUsers()
    } catch { /* ignore */ }
  }

  async function handleDisable(u: UserDetails) {
    if (!confirm(`Disable ${u.email}?`)) return
    try {
      await users.disable(u.id)
      fetchUsers()
    } catch { /* ignore */ }
  }

  async function handleEnable(u: UserDetails) {
    try {
      await users.enable(u.id)
      fetchUsers()
    } catch { /* ignore */ }
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Search Users</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Input
            placeholder="First name"
            value={search.firstName}
            onChange={(e) => handleSearch('firstName', e.target.value)}
          />
          <Input
            placeholder="Last name"
            value={search.lastName}
            onChange={(e) => handleSearch('lastName', e.target.value)}
          />
          <Input
            placeholder="Email"
            value={search.email}
            onChange={(e) => handleSearch('email', e.target.value)}
          />
        </div>
        <Button size="sm" className="mt-3" onClick={fetchUsers}>
          Search
        </Button>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">Users ({userList.length})</h3>
        </div>
        {loading ? (
          <div className="p-6 text-center text-gray-500">Loading...</div>
        ) : userList.length === 0 ? (
          <div className="p-6 text-center text-gray-500">No users found</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 text-left">
                <tr>
                  <th className="px-6 py-3 font-medium text-gray-500">Name</th>
                  <th className="px-6 py-3 font-medium text-gray-500">Email</th>
                  <th className="px-6 py-3 font-medium text-gray-500">Roles</th>
                  <th className="px-6 py-3 font-medium text-gray-500">Status</th>
                  <th className="px-6 py-3 font-medium text-gray-500">2FA</th>
                  <th className="px-6 py-3 font-medium text-gray-500">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {userList.map(u => (
                  <tr key={u.id} className="hover:bg-gray-50">
                    <td className="px-6 py-3">
                      <Link to={`/admin/users/${u.id}`} className="font-medium text-gray-900 hover:text-indigo-600">
                        {u.firstName} {u.lastName}
                      </Link>
                    </td>
                    <td className="px-6 py-3 text-gray-600">{u.email}</td>
                    <td className="px-6 py-3">
                      <div className="flex flex-wrap gap-1">
                        {u.roles.map((r, i) => {
                          let roleName = r
                          let roleDesc = ''
                          try {
                            const parsed = JSON.parse(r)
                            roleName = parsed.name || r
                            roleDesc = parsed.description || ''
                          } catch { /* not JSON, use as-is */ }
                          return (
                            <div key={i} className="relative group">
                              <Badge variant="info">{roleName}</Badge>
                              {roleDesc && (
                                <div className="absolute z-10 bottom-full left-1/2 -translate-x-1/2 mb-2 w-48 bg-gray-900 text-white text-xs rounded-lg px-3 py-2 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none shadow-lg">
                                  <p className="text-gray-300">{roleDesc}</p>
                                </div>
                              )}
                            </div>
                          )
                        })}
                      </div>
                    </td>
                    <td className="px-6 py-3">
                      {u.locked && <Badge variant="warning">Locked</Badge>}
                      {u.locked && ' '}
                      {u.locked && <Badge variant="danger">Disabled</Badge>}
                      {!u.locked && <Badge variant="success">Active</Badge>}
                    </td>
                    <td className="px-6 py-3">
                      {u.twoFAEnabled ? <Badge variant="success">ON</Badge> : <Badge>OFF</Badge>}
                    </td>
                    <td className="px-6 py-3">
                      <div className="flex gap-2">
                        <Button variant="warning" size="sm" onClick={() => handleToggleLock(u)} disabled={u.id === currentUser?.id}>
                          {u.locked ? 'Unlock' : 'Lock'}
                        </Button>
                        <Button variant="danger" size="sm" onClick={() => u.disabled ? handleEnable(u) : handleDisable(u)} disabled={u.id === currentUser?.id}>
                          {u.disabled ? 'Enable' : 'Disable'}
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
