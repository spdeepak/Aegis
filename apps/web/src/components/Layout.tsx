import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'

const navItems = [
  { label: 'Dashboard', path: '/', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' },
  { label: 'Profile', path: '/profile', icon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' },
]

const adminItems = [
  { label: 'Users', path: '/admin/users', permission: 'users:read', icon: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' },
  { label: 'Roles', path: '/admin/roles', permission: 'roles:read', icon: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z' },
  { label: 'Permissions', path: '/admin/permissions', permission: 'permissions:read', icon: 'M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z' },
]

export function Layout() {
  const { user, hasPermission, hasRole, logout } = useAuth()
  const location = useLocation()
  const navigate = useNavigate()

  const canSeeAdmin = hasRole('super_admin') || adminItems.some(i => hasPermission(i.permission))

  function handleLogout() {
    logout()
    navigate('/login')
  }

  return (
    <div className="flex min-h-screen bg-gray-50">
      <aside className="w-64 bg-gray-900 text-white flex flex-col min-h-screen shrink-0">
        <div className="p-4 border-b border-gray-800">
          <h1 className="text-xl font-bold tracking-tight">Aegis</h1>
          {user && (
            <p className="text-xs text-gray-400 mt-1 truncate">{user.email}</p>
          )}
        </div>

        <nav className="flex-1 p-3 space-y-1">
          {navItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className={`flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                location.pathname === item.path
                  ? 'bg-gray-800 text-white'
                  : 'text-gray-400 hover:bg-gray-800/50 hover:text-white'
              }`}
            >
              <svg className="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d={item.icon} />
              </svg>
              {item.label}
            </Link>
          ))}

          {canSeeAdmin && (
            <>
              <div className="pt-4 pb-2 px-3">
                <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Admin</p>
              </div>
              {adminItems.map((item) => {
                const visible = hasRole('super_admin') || hasPermission(item.permission)
                if (!visible) return null
                return (
                  <Link
                    key={item.path}
                    to={item.path}
                    className={`flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                      location.pathname.startsWith(item.path)
                        ? 'bg-gray-800 text-white'
                        : 'text-gray-400 hover:bg-gray-800/50 hover:text-white'
                    }`}
                  >
                    <svg className="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d={item.icon} />
                    </svg>
                    {item.label}
                  </Link>
                )
              })}
            </>
          )}
        </nav>

        <div className="p-3 border-t border-gray-800">
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-gray-400 hover:bg-gray-800/50 hover:text-white w-full transition-colors"
          >
            <svg className="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            Sign out
          </button>
        </div>
      </aside>

      <main className="flex-1 overflow-auto">
        <header className="bg-white border-b border-gray-200 px-6 py-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-800">
              {getPageTitle(location.pathname)}
            </h2>
            <div className="flex items-center gap-3">
              <div className="text-right">
                <p className="text-sm font-medium text-gray-700">{user?.firstName} {user?.lastName}</p>
                <div className="flex gap-1 mt-0.5">
                  {user?.roles.map(role => (
                    <span key={role} className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-indigo-100 text-indigo-800">
                      {role}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </header>
        <div className="p-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}

function getPageTitle(pathname: string): string {
  if (pathname === '/') return 'Dashboard'
  if (pathname === '/profile') return 'Profile'
  if (pathname === '/admin/users') return 'Users'
  if (pathname.match(/^\/admin\/users\/\d+$/)) return 'User Details'
  if (pathname === '/admin/roles') return 'Roles'
  if (pathname.match(/^\/admin\/roles\/\d+$/)) return 'Role Details'
  if (pathname === '/admin/permissions') return 'Permissions'
  return 'Aegis'
}
