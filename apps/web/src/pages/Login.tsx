import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { auth, setTokens } from '../lib/api'
import { useAuth } from '../contexts/AuthContext'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'

export function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const { login } = useAuth()

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const result = await auth.login(email, password)
      if ('accessToken' in result) {
        setTokens(result.accessToken, result.refreshToken)
        login(result.accessToken, result.refreshToken)
        navigate('/')
      } else if (result.type === '2fa') {
        navigate('/2fa', { state: { tempToken: result.temp_token } })
      }
    } catch (err: unknown) {
      const msg = (err && typeof err === 'object' && 'description' in err)
        ? String((err as { description: string }).description)
        : 'Login failed'
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
          <p className="text-gray-500 mt-2">Sign in to your account</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 space-y-4">
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3">
              {error}
            </div>
          )}

          <Input
            label="Email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="admin@localhost"
            required
          />

          <Input
            label="Password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Enter your password"
            required
          />

          <Button type="submit" className="w-full" disabled={loading}>
            {loading ? 'Signing in...' : 'Sign in'}
          </Button>

          <p className="text-center text-sm text-gray-500">
            Don't have an account?{' '}
            <Link to="/signup" className="text-indigo-600 hover:text-indigo-500 font-medium">
              Sign up
            </Link>
          </p>
        </form>
      </div>
    </div>
  )
}
