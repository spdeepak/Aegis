import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { auth } from '../lib/api'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'

export function Signup() {
  const [email, setEmail] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [password, setPassword] = useState('')
  const [twoFAEnabled, setTwoFAEnabled] = useState(false)
  const [qrImage, setQrImage] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const result = await auth.signup({ email, firstName, lastName, password, twoFAEnabled })
      if (result && 'qr_image' in result) {
        setQrImage(result.qr_image)
      } else {
        navigate('/login')
      }
    } catch (err: unknown) {
      const msg = (err && typeof err === 'object' && 'description' in err)
        ? String((err as { description: string }).description)
        : 'Signup failed'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  if (qrImage) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
        <div className="w-full max-w-md text-center">
          <h1 className="text-3xl font-bold text-gray-900 mb-4">Aegis</h1>
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-semibold mb-4">Scan QR Code</h2>
            <p className="text-sm text-gray-500 mb-4">
              Scan this QR code with your authenticator app to complete 2FA setup.
            </p>
            <img src={qrImage} alt="2FA QR Code" className="mx-auto mb-4" />
            <Link to="/login" className="text-indigo-600 hover:text-indigo-500 font-medium text-sm">
              Go to login
            </Link>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Aegis</h1>
          <p className="text-gray-500 mt-2">Create your account</p>
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
            required
          />

          <div className="grid grid-cols-2 gap-4">
            <Input
              label="First Name"
              value={firstName}
              onChange={(e) => setFirstName(e.target.value)}
              required
            />
            <Input
              label="Last Name"
              value={lastName}
              onChange={(e) => setLastName(e.target.value)}
              required
            />
          </div>

          <Input
            label="Password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Min 8 chars, upper, lower, number, special"
            required
          />

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="twoFA"
              checked={twoFAEnabled}
              onChange={(e) => setTwoFAEnabled(e.target.checked)}
              className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
            />
            <label htmlFor="twoFA" className="text-sm text-gray-700">Enable 2FA</label>
          </div>

          <Button type="submit" className="w-full" disabled={loading}>
            {loading ? 'Creating account...' : 'Create account'}
          </Button>

          <p className="text-center text-sm text-gray-500">
            Already have an account?{' '}
            <Link to="/login" className="text-indigo-600 hover:text-indigo-500 font-medium">
              Sign in
            </Link>
          </p>
        </form>
      </div>
    </div>
  )
}
