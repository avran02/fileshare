package models

import (
	"time"
)

type Token struct {
	ID        uint64
	UserID    string
	Type      string
	Token     string
	ExpiresAt time.Time
}
