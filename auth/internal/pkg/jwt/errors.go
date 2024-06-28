package jwt

import "errors"

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrWrongTokenType = errors.New("wrong token type")
	ErrExpiredToken   = errors.New("token expired")
	ErrEmptyToken     = errors.New("token is empty")
)
