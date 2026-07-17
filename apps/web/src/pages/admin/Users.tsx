import { useState, useEffect, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { users, type UserDetails } from '../../lib/api'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'

export function UsersPage() {
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
                        {u.roles.map(r => (
                          <Badge key={r} variant="info">{r}</Badge>
                        ))}
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
                        <Button variant="ghost" size="sm" onClick={() => handleToggleLock(u)}>
                          {u.locked ? 'Unlock' : 'Lock'}
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => u.disabled ? handleEnable(u) : handleDisable(u)}>
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
