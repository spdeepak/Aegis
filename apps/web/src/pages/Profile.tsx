import { useState, useEffect, type FormEvent } from 'react'
import { auth, type Session } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'

export function Profile() {
  const { user } = useAuth()
  const [sessions, setSessions] = useState<Session[]>([])
  const [loadingSessions, setLoadingSessions] = useState(true)
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [passwordMsg, setPasswordMsg] = useState('')
  const [passwordErr, setPasswordErr] = useState('')
  const [changingPw, setChangingPw] = useState(false)

  useEffect(() => {
    auth.getSessions().then(setSessions).catch(() => {}).finally(() => setLoadingSessions(false))
  }, [])

  async function handleChangePassword(e: FormEvent) {
    e.preventDefault()
    setPasswordMsg('')
    setPasswordErr('')
    setChangingPw(true)
    try {
      await auth.changePassword(oldPassword, newPassword)
      setPasswordMsg('Password changed successfully')
      setOldPassword('')
      setNewPassword('')
    } catch (err: unknown) {
      const msg = (err && typeof err === 'object' && 'description' in err)
        ? String((err as { description: string }).description)
        : 'Failed to change password'
      setPasswordErr(msg)
    } finally {
      setChangingPw(false)
    }
  }

  async function handleRevokeAll() {
    if (!confirm('Revoke all sessions? This will sign you out everywhere.')) return
    try {
      await auth.revokeAllSessions()
      setSessions([])
    } catch { /* ignore */ }
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Account</h2>
        <dl className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <dt className="text-gray-500">Name</dt>
            <dd className="font-medium text-gray-900">{user?.firstName} {user?.lastName}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Email</dt>
            <dd className="font-medium text-gray-900">{user?.email}</dd>
          </div>
          <div>
            <dt className="text-gray-500">User ID</dt>
            <dd className="font-medium text-gray-900">{user?.id}</dd>
          </div>
        </dl>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Change Password</h2>
        <form onSubmit={handleChangePassword} className="space-y-4">
          {passwordMsg && (
            <div className="bg-green-50 border border-green-200 text-green-700 text-sm rounded-lg px-4 py-3">{passwordMsg}</div>
          )}
          {passwordErr && (
            <div className="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3">{passwordErr}</div>
          )}
          <Input
            label="Current Password"
            type="password"
            value={oldPassword}
            onChange={(e) => setOldPassword(e.target.value)}
            required
          />
          <Input
            label="New Password"
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            required
          />
          <Button type="submit" disabled={changingPw}>
            {changingPw ? 'Changing...' : 'Change Password'}
          </Button>
        </form>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900">Active Sessions</h2>
          {sessions.length > 0 && (
            <Button variant="danger" size="sm" onClick={handleRevokeAll}>
              Revoke All
            </Button>
          )}
        </div>
        {loadingSessions ? (
          <p className="text-sm text-gray-500">Loading sessions...</p>
        ) : sessions.length === 0 ? (
          <p className="text-sm text-gray-500">No active sessions</p>
        ) : (
          <div className="space-y-3">
            {sessions.map((s, i) => (
              <div key={i} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <div className="text-sm">
                  <p className="font-medium text-gray-900">{s.createdBy || 'Unknown'}</p>
                  <p className="text-gray-500 text-xs mt-0.5">{s.userAgent}</p>
                  <p className="text-gray-400 text-xs">IP: {s.ipAddress}</p>
                </div>
                <div className="text-right text-xs text-gray-400">
                  <p>Issued: {new Date(s.issuedAt).toLocaleDateString()}</p>
                  <p>Expires: {new Date(s.expiresAt).toLocaleDateString()}</p>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
