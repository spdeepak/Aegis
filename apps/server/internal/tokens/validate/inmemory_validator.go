package tokens

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spdeepak/aegis/server/api"
	"github.com/spdeepak/aegis/server/internal/tokens"
	"github.com/spdeepak/aegis/server/pkg/ttlcache"
)

type (
	inMemoryValidator struct {
		secret          []byte
		issuer          string
		tokenRepository tokens.Querier
		cache           *ttlcache.Cache
	}
)

func NewInMemoryValidator(tokenRepository tokens.Querier, secret []byte, issuer string, cache *ttlcache.Cache) Validator {
	//refreshTokens, err := tokenRepository.ListRevokedValidRefreshTokens(context.Background())
	//if err != nil {
	//	slog.Error("error listing revoked refresh tokens: ", err)
	//	return nil
	//}

	return &inMemoryValidator{
		secret:          secret,
		issuer:          issuer,
		tokenRepository: tokenRepository,
		cache:           cache,
	}
}

func (i *inMemoryValidator) ValidateRefreshToken(ctx context.Context, clientIP string, params api.RefreshParams, refreshToken string) (jwt.MapClaims, error) {
	//TODO implement me
	panic("implement me")
}
