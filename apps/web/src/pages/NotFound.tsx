import { Link } from 'react-router-dom'

export function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <h1 className="text-6xl font-bold text-gray-300">404</h1>
        <p className="text-gray-500 mt-2">Page not found</p>
        <Link to="/" className="text-indigo-600 hover:text-indigo-500 font-medium text-sm mt-4 inline-block">
          Go to dashboard
        </Link>
      </div>
    </div>
  )
}
