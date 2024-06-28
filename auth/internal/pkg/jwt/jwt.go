package jwt

import (
	"fmt"
	"time"

	"github.com/avran02/auth/internal/config"
	"github.com/avran02/auth/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type JwtGenerator interface {
	Generate(id string, isAccess bool) (models.Token, error)
	Validate(token string) (userId string, isRefresh bool, err error) //nolint
}

type jwtToken struct {
	config.JWT
}

func (j *jwtToken) Generate(id string, isAccess bool) (models.Token, error) {
	var tokenType string
	claims := newClaims(id, isAccess, j.Exp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return models.Token{}, fmt.Errorf("failed to sign token: %w", err)
	}

	if isAccess {
		tokenType = accessTokenType
	} else {
		tokenType = refreshTokenType
	}

	return models.Token{
		UserID:    id,
		Type:      tokenType,
		Token:     signedToken,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

func (j *jwtToken) Validate(token string) (userID string, isAccess bool, err error) {
	if token == "" {
		return "", false, ErrEmptyToken
	}

	parsedToken, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.Secret), nil
	})
	if err != nil {
		return "", false, fmt.Errorf("failed to parse token: %w", err)
	}

	if !parsedToken.Valid {
		return "", false, ErrInvalidToken
	}

	claims, ok := parsedToken.Claims.(*claims)
	if !ok {
		return "", false, ErrInvalidToken
	}

	if claims.Type != refreshTokenType && claims.Type != accessTokenType {
		return "", false, ErrInvalidToken
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return "", false, ErrExpiredToken
	}

	return claims.Subject, claims.Type == accessTokenType, nil
}

func New(config config.JWT) JwtGenerator {
	return &jwtToken{
		JWT: config,
	}
}
