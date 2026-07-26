package tokens

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spdeepak/aegis/server/api"
)

type (
	Validator interface {
		ValidateRefreshToken(ctx context.Context, clientIP string, params api.RefreshParams, refreshToken string) (jwt.MapClaims, error)
	}
)
