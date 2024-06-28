package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var oneWeekDuration = 7 * 24 * time.Hour

const (
	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

type claims struct { // TODO: add roles
	Type string `json:"type"`
	jwt.RegisteredClaims
}

func newClaims(userId string, isAccess bool, exp int) claims { //nolint
	var expTime time.Time
	var tokenType string

	if isAccess {
		expTime = time.Now().Add(time.Duration(exp) * time.Second)
		tokenType = accessTokenType
	} else {
		expTime = time.Now().Add(oneWeekDuration)
		tokenType = refreshTokenType
	}

	return claims{
		Type: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userId,
			ExpiresAt: jwt.NewNumericDate(expTime),
		},
	}
}
