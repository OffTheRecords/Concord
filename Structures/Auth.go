package Structures

import "time"

type AuthTokens struct {
	AccessToken   string
	RefreshToken  string
	RefreshID     string
	AccessExpiry  time.Time
	RefreshExpiry time.Time
}
