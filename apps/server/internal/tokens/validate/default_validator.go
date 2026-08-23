package tokens

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/spdeepak/aegis/server/api"
	httperror "github.com/spdeepak/aegis/server/internal/error"
	"github.com/spdeepak/aegis/server/internal/tokens"
	pkgtime "github.com/spdeepak/aegis/server/pkg/time"
)

type (
	defaultValidator struct {
		secret          []byte
		issuer          string
		tokenRepository tokens.Querier
	}
)

func NewDefaultValidator(tokenRepository tokens.Querier, secret []byte, issuer string) Validator {
	return &defaultValidator{
		tokenRepository: tokenRepository,
		secret:          secret,
		issuer:          issuer,
	}
}

func (s *defaultValidator) ValidateRefreshToken(ctx context.Context, clientIP string, params api.RefreshParams, refreshToken string) (jwt.MapClaims, error) {
	claims, err := s.verifyToken(refreshToken)
	if err != nil {
		return nil, err
	}
	refreshValidParams := tokens.IsRefreshValidParams{
		RefreshToken: hash(refreshToken),
		IpAddress:    clientIP,
		UserAgent:    params.UserAgent,
		DeviceName:   "",
	}
	res, err := s.tokenRepository.IsRefreshValid(ctx, refreshValidParams)
	if err != nil {
		return nil, httperror.NewWithMetadata(httperror.RefreshTokenRevoked, err.Error())
	} else if res != 1 { //value of res should be 1 to be valid
		return nil, httperror.New(httperror.RefreshTokenRevoked)
	}
	return claims, nil
}

func (s *defaultValidator) verifyToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, httperror.NewWithMetadata(httperror.UndefinedErrorCode, fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, httperror.NewWithMetadata(httperror.Unauthorized, err.Error())
	} else if !token.Valid {
		return nil, httperror.NewWithMetadata(httperror.Unauthorized, "Invalid Refresh Token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		if expTime, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(expTime), 0)
			if expirationTime.Before(pkgtime.Now()) {
				return nil, httperror.New(httperror.ExpiredRefreshToken)
			}
		} else {
			return nil, httperror.NewWithMetadata(httperror.UndefinedErrorCode, "invalid claims")
		}
	} else if !ok || !token.Valid {
		return nil, httperror.NewWithMetadata(httperror.UndefinedErrorCode, "invalid claims")
	}

	return claims, nil
}

func hash(anything string) string {
	h := sha256.New()
	h.Write([]byte(anything))
	return hex.EncodeToString(h.Sum(nil))
}
