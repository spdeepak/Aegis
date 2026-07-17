import { Routes, Route } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import { ProtectedRoute } from './components/ProtectedRoute'
import { AdminRoute } from './components/AdminRoute'
import { Layout } from './components/Layout'
import { Login } from './pages/Login'
import { Signup } from './pages/Signup'
import { Login2FA } from './pages/Login2FA'
import { Dashboard } from './pages/Dashboard'
import { Profile } from './pages/Profile'
import { NotFound } from './pages/NotFound'
import { UsersPage } from './pages/admin/Users'
import { UserDetail } from './pages/admin/UserDetail'
import { RolesPage } from './pages/admin/Roles'
import { RoleDetail } from './pages/admin/RoleDetail'
import { PermissionsPage } from './pages/admin/Permissions'

export default function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/2fa" element={<Login2FA />} />

        <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/profile" element={<Profile />} />

          <Route path="/admin/users" element={
            <AdminRoute requirePermission="users:read"><UsersPage /></AdminRoute>
          } />
          <Route path="/admin/users/:id" element={
            <AdminRoute requirePermission="users:read"><UserDetail /></AdminRoute>
          } />
          <Route path="/admin/roles" element={
            <AdminRoute requirePermission="roles:read"><RolesPage /></AdminRoute>
          } />
          <Route path="/admin/roles/:id" element={
            <AdminRoute requirePermission="roles:read"><RoleDetail /></AdminRoute>
          } />
          <Route path="/admin/permissions" element={
            <AdminRoute requirePermission="permissions:read"><PermissionsPage /></AdminRoute>
          } />
        </Route>

        <Route path="*" element={<NotFound />} />
      </Routes>
    </AuthProvider>
  )
}
