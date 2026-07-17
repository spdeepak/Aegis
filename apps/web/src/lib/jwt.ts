interface JwtPayload {
  name?: string
  email: string
  first_name?: string
  last_name?: string
  typ: string
  auth_level?: string
  roles?: string[]
  permissions?: string[]
  sub: string
  iss?: string
  exp?: number
  iat?: number
  jti?: string
}

export interface DecodedUser {
  id: number
  email: string
  firstName: string
  lastName: string
  roles: string[]
  permissions: string[]
  expiresAt: number
}

function base64UrlDecode(str: string): string {
  const base64 = str.replace(/-/g, '+').replace(/_/g, '/')
  const padded = base64 + '='.repeat((4 - (base64.length % 4)) % 4)
  return atob(padded)
}

export function decodeToken(token: string): JwtPayload | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return null
    const payload = JSON.parse(base64UrlDecode(parts[1]))
    return payload as JwtPayload
  } catch {
    return null
  }
}

export function getUserFromToken(token: string): DecodedUser | null {
  const payload = decodeToken(token)
  if (!payload) return null
  if (payload.typ !== 'Bearer') return null

  return {
    id: parseInt(payload.sub, 10),
    email: payload.email,
    firstName: payload.first_name || '',
    lastName: payload.last_name || '',
    roles: payload.roles || [],
    permissions: payload.permissions || [],
    expiresAt: payload.exp || 0,
  }
}

export function isTokenExpired(token: string): boolean {
  const payload = decodeToken(token)
  if (!payload || !payload.exp) return true
  return Date.now() >= payload.exp * 1000
}
