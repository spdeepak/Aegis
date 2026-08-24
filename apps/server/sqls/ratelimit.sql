-- name: RecordAuthFailure :one
INSERT INTO auth_rate_limits (key, failures, window_start, locked_until, updated_at)
VALUES (sqlc.arg('key'), 1, now(), NULL, now())
ON CONFLICT (key) DO UPDATE
SET failures = CASE
                   WHEN auth_rate_limits.locked_until > now() THEN auth_rate_limits.failures
                   -- window elapsed: restart the failure counter
                   WHEN auth_rate_limits.updated_at < now() - sqlc.arg('window')::interval THEN 1
                   ELSE auth_rate_limits.failures + 1 END,
    locked_until = CASE
                       WHEN auth_rate_limits.locked_until > now() THEN auth_rate_limits.locked_until
                       WHEN (CASE
                                 WHEN auth_rate_limits.updated_at < now() - sqlc.arg('window')::interval THEN 1
                                 ELSE auth_rate_limits.failures + 1 END) >= sqlc.arg('max_attempts')::int
                           THEN now() + sqlc.arg('lockout')::interval
                       ELSE NULL END,
    window_start = CASE
                       WHEN auth_rate_limits.updated_at < now() - sqlc.arg('window')::interval THEN now()
                       ELSE auth_rate_limits.window_start END,
    updated_at = now()
-- Returns whether the key is now locked out (handles the NULL locked_until case).
RETURNING (COALESCE(auth_rate_limits.locked_until > now(), false))::boolean AS locked;

-- name: IsLocked :one
SELECT EXISTS (
    SELECT 1
    FROM auth_rate_limits
    WHERE key = sqlc.arg('key')
      AND locked_until > now()
) AS locked;

-- name: ResetAuthFailure :exec
UPDATE auth_rate_limits
SET failures = 0,
    locked_until = NULL,
    updated_at = now()
WHERE key = sqlc.arg('key');
