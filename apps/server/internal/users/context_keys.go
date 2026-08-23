package users

type contextKey string

const (
	CtxKeyUserID contextKey = "User-ID"
	CtxKeyUserIP contextKey = "user-ip"
)
