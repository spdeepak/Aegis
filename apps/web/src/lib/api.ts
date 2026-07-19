import { isTokenExpired } from './jwt'

const ACCESS_TOKEN_KEY = 'aegis_access_token'
const REFRESH_TOKEN_KEY = 'aegis_refresh_token'

function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

function setTokens(accessToken: string, refreshToken: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
}

function clearTokens(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
}

let isRefreshing = false
let refreshPromise: Promise<string> | null = null

async function refreshAccessToken(): Promise<string> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    clearTokens()
    window.location.href = '/login'
    throw new Error('No refresh token')
  }

  if (isRefreshing && refreshPromise) {
    return refreshPromise
  }

  isRefreshing = true
  refreshPromise = (async () => {
    try {
      const res = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'User-Agent': navigator.userAgent,
          'x-login-source': 'web',
        },
        body: JSON.stringify({ refreshToken }),
      })

      if (!res.ok) {
        if (res.status === 401 || res.status === 403) {
          clearTokens()
          window.location.href = '/login'
        }
        throw new Error('Refresh failed')
      }

      const data = await res.json()
      setTokens(data.accessToken, data.refreshToken)
      return data.accessToken as string
    } finally {
      isRefreshing = false
      refreshPromise = null
    }
  })()

  return refreshPromise
}

async function ensureValidToken(): Promise<string | null> {
  const token = getAccessToken()
  if (!token) return null
  if (!isTokenExpired(token)) return token
  return refreshAccessToken()
}

interface RequestOptions extends Omit<RequestInit, 'method' | 'body'> {
  method?: string
  body?: unknown
  params?: Record<string, string | number | undefined>
}

async function request<T>(url: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, params, headers: extraHeaders, ...rest } = options

  let fullUrl = url
  if (params) {
    const searchParams = new URLSearchParams()
    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined && value !== '') {
        searchParams.set(key, String(value))
      }
    }
    const qs = searchParams.toString()
    if (qs) fullUrl += `?${qs}`
  }

  const headers: Record<string, string> = {
    'User-Agent': navigator.userAgent,
    'x-login-source': 'web',
    ...(extraHeaders as Record<string, string>),
  }

  const token = await ensureValidToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  if (body && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }

  const res = await fetch(fullUrl, {
    method,
    headers,
    body: body ? (body instanceof FormData ? body : JSON.stringify(body)) : undefined,
    ...rest,
  })

  if (!res.ok) {
    const errorBody = await res.json().catch(() => ({}))
    throw { status: res.status, ...errorBody }
  }

  if (res.status === 204 || res.headers.get('content-length') === '0') {
    return undefined as T
  }

  return res.json()
}

// ─── Auth ───────────────────────────────────────────────────────

export interface LoginSuccess {
  accessToken: string
  refreshToken: string
}

export interface LoginRequires2FA {
  type: '2fa'
  temp_token: string
}

export interface SignUpWith2FAResponse {
  qr_image: string
  secret: string
}

export interface Session {
  issuedAt: string
  expiresAt: string
  ipAddress: string
  userAgent: string
  createdBy: string
}

export const auth = {
  login: (email: string, password: string) =>
    request<LoginSuccess | LoginRequires2FA>('/api/v1/auth/login', {
      method: 'POST',
      body: { email, password },
    }),

  login2FA: (twoFACode: string) =>
    request<LoginSuccess>('/api/v1/auth/2fa/login', {
      method: 'POST',
      body: { twoFACode },
    }),

  signup: (data: {
    email: string
    firstName: string
    lastName: string
    password: string
    twoFAEnabled: boolean
  }) =>
    request<SignUpWith2FAResponse | void>('/api/v1/auth/signup', {
      method: 'POST',
      body: data,
    }),

  changePassword: (oldPassword: string, newPassword: string, twoFACode?: string) =>
    request<void>('/api/v1/auth/password', {
      method: 'POST',
      body: { oldPassword, newPassword, twoFACode },
    }),

  refresh: (refreshToken: string) =>
    request<LoginSuccess>('/api/v1/auth/refresh', {
      method: 'POST',
      body: { refreshToken },
    }),

  getSessions: () => request<Session[]>('/api/v1/auth/sessions'),

  revokeSession: (refreshToken: string) =>
    request<void>('/api/v1/auth/sessions/current', {
      method: 'DELETE',
      body: { refreshToken },
    }),

  revokeAllSessions: () => request<void>('/api/v1/auth/sessions', { method: 'DELETE' }),

  setup2FA: () => request<{ qr_image: string; secret: string }>('/api/v1/auth/2fa/setup', { method: 'POST' }),

  remove2FA: (twoFACode: string) =>
    request<void>('/api/v1/auth/2fa', {
      method: 'DELETE',
      body: { twoFACode },
    }),
}

// ─── Users (Admin) ──────────────────────────────────────────────

export interface UserDetails {
  id: number
  email: string
  firstName: string
  lastName: string
  locked: boolean
  disabled: boolean
  twoFAEnabled: boolean
  createdAt: string
  updatedAt: string
  roles: string[]
  permissions: string[]
}

export interface UserWithRoles {
  id: number
  email: string
  firstName: string
  lastName: string
  roles: string[]
  permissions: string[]
}

export const users = {
  list: (params?: { firstName?: string; lastName?: string; email?: string; size?: number; page?: number }) =>
    request<UserDetails[]>('/api/v1/users', { params }),

  getRoles: (id: number) => request<UserWithRoles>(`/api/v1/users/${id}/roles`),

  assignRoles: (id: number, roleIds: number[]) =>
    request<void>(`/api/v1/users/${id}/roles`, {
      method: 'POST',
      body: { roles: roleIds },
    }),

  removeRole: (userId: number, roleId: number) =>
    request<void>(`/api/v1/users/${userId}/roles/${roleId}`, { method: 'DELETE' }),

  lock: (id: number) => request<void>(`/api/v1/users/${id}/lock`, { method: 'POST' }),

  unlock: (id: number) => request<void>(`/api/v1/users/${id}/lock`, { method: 'DELETE' }),

  disable: (id: number) => request<void>(`/api/v1/users/${id}/disable`, { method: 'POST' }),

  enable: (id: number) => request<void>(`/api/v1/users/${id}/enable`, { method: 'POST' }),
}

// ─── Roles ──────────────────────────────────────────────────────

export interface RoleResponse {
  id: number
  name: string
  description: string
  createdAt: string
  createdBy: string
  updatedAt: string
  updatedBy: string
}

export interface RolesAndPermission {
  roles: {
    id: number
    name: string
    description: string
    createdAt: string
    createdBy: string
    updatedAt: string
    updatedBy: string
    permissions: PermissionResponse[]
  }
}

export const roles = {
  list: () => request<RoleResponse[]>('/api/v1/access-control/roles'),

  get: (id: number) => request<RoleResponse>(`/api/v1/access-control/roles/${id}`),

  create: (data: { name: string; description: string }) =>
    request<RoleResponse>('/api/v1/access-control/roles', {
      method: 'POST',
      body: data,
    }),

  update: (id: number, data: { name?: string; description?: string }) =>
    request<RoleResponse>(`/api/v1/access-control/roles/${id}`, {
      method: 'PATCH',
      body: data,
    }),

  delete: (id: number) =>
    request<void>(`/api/v1/access-control/roles/${id}`, { method: 'DELETE' }),

  listAndPermissions: () => request<RolesAndPermission[]>('/api/v1/access-control/roles/permissions'),

  assignPermissions: (roleId: number, permissionIds: number[]) =>
    request<void>(`/api/v1/access-control/roles/${roleId}/permissions`, {
      method: 'POST',
      body: { ids: permissionIds },
    }),

  unassignPermission: (roleId: number, permissionId: number) =>
    request<void>(`/api/v1/access-control/roles/${roleId}/permissions/${permissionId}`, {
      method: 'DELETE',
    }),
}

// ─── Permissions ────────────────────────────────────────────────

export interface PermissionResponse {
  id: number
  name: string
  description: string
  createdAt: string
  createdBy: string
  updatedAt: string
  updatedBy: string
}

export const permissions = {
  list: () => request<PermissionResponse[]>('/api/v1/access-control/permissions'),

  get: (id: number) => request<PermissionResponse>(`/api/v1/access-control/permissions/${id}`),

  create: (data: { name: string; description: string }) =>
    request<PermissionResponse>('/api/v1/access-control/permissions', {
      method: 'POST',
      body: data,
    }),

  update: (id: number, data: { name?: string; description?: string }) =>
    request<PermissionResponse>(`/api/v1/access-control/permissions/${id}`, {
      method: 'PATCH',
      body: data,
    }),

  delete: (id: number) =>
    request<void>(`/api/v1/access-control/permissions/${id}`, { method: 'DELETE' }),
}

export { setTokens, clearTokens, getAccessToken, getRefreshToken, refreshAccessToken }
