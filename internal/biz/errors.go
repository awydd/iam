package biz

import (
	"errors"
)

// user
var (
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrAccountNotActive    = errors.New("account is not active")
	ErrAccountLocked       = errors.New("account is temporarily locked due to too many failed login attempts")
	ErrRefreshTokenInvalid = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token expired or revoked")
)
