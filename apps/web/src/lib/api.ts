import { isTokenExpired } from './jwt'
import type {
  RoleResponse,
  PermissionResponse,
  UserDetails,
  UserWithRoles,
  GetAllSessionResponse,
  LoginSuccessWithJWT,
  LoginRequires2FA,
  TwoFAResponse,
  SignUpWith2FAResponse,
  RolesAndPermissionResponse,
  GetListOfUsersParams,
} from './api.gen'

// ─── Token Management ──────────────────────────────────────────

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
        throw { status: res.status }
      }

      const data: LoginSuccessWithJWT = await res.json()
      setTokens(data.accessToken, data.refreshToken)
      return data.accessToken
    } catch (err) {
      clearTokens()
      window.location.href = '/login'
      throw err
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

// ─── Request Helper ─────────────────────────────────────────────

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

export const auth = {
  login: (email: string, password: string) =>
    request<LoginSuccessWithJWT | LoginRequires2FA>('/api/v1/auth/login', {
      method: 'POST',
      body: { email, password },
    }),

  login2FA: (twoFACode: string) =>
    request<LoginSuccessWithJWT>('/api/v1/auth/2fa/login', {
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
    request<LoginSuccessWithJWT>('/api/v1/auth/refresh', {
      method: 'POST',
      body: { refreshToken },
    }),

  getSessions: () => request<GetAllSessionResponse[]>('/api/v1/auth/sessions'),

  revokeSession: (refreshToken: string) =>
    request<void>('/api/v1/auth/sessions/current', {
      method: 'DELETE',
      body: { refreshToken },
    }),

  revokeAllSessions: () => request<void>('/api/v1/auth/sessions', { method: 'DELETE' }),

  setup2FA: () => request<TwoFAResponse>('/api/v1/auth/2fa/setup', { method: 'POST' }),

  remove2FA: (twoFACode: string) =>
    request<void>('/api/v1/auth/2fa', {
      method: 'DELETE',
      body: { twoFACode },
    }),
}

// ─── Users (Admin) ──────────────────────────────────────────────

export const users = {
  list: (params?: GetListOfUsersParams) =>
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

  listAndPermissions: () => request<RolesAndPermissionResponse[]>('/api/v1/access-control/roles/permissions'),

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

// ─── Re-exports ─────────────────────────────────────────────────

export { setTokens, clearTokens, getAccessToken, getRefreshToken, refreshAccessToken }
export type { RoleResponse, PermissionResponse, UserDetails, UserWithRoles, GetAllSessionResponse as Session }
