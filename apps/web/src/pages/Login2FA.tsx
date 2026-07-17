import { useState, type FormEvent } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { auth, setTokens } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'

export function Login2FA() {
  const [code, setCode] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const location = useLocation()
  const navigate = useNavigate()
  const { login } = useAuth()

  const tempToken = (location.state as { tempToken?: string })?.tempToken

  if (!tempToken) {
    navigate('/login')
    return null
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const result = await auth.login2FA(code)
      setTokens(result.accessToken, result.refreshToken)
      login(result.accessToken, result.refreshToken)
      navigate('/')
    } catch (err: unknown) {
      const msg = (err && typeof err === 'object' && 'description' in err)
        ? String((err as { description: string }).description)
        : 'Invalid code'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Aegis</h1>
          <p className="text-gray-500 mt-2">Enter your 2FA code</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 space-y-4">
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3">
              {error}
            </div>
          )}

          <Input
            label="6-digit code"
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
            placeholder="000000"
            maxLength={6}
            required
          />

          <Button type="submit" className="w-full" disabled={loading || code.length !== 6}>
            {loading ? 'Verifying...' : 'Verify'}
          </Button>
        </form>
      </div>
    </div>
  )
}
