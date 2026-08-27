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
	ErrUserNotFound        = errors.New("user not found")
	ErrApplicationDisabled = errors.New("application is disabled")
)

// application
var (
	ErrClientIDRequired  = errors.New("client_id is required")
	ErrClientSecretWrong = errors.New("client secret is invalid")
)

// oauth
var (
	ErrClientNotFound     = errors.New("client not found")
	ErrClientDisabled     = errors.New("client application is disabled")
	ErrRedirectURIInvalid = errors.New("redirect_uri is not registered for this client")
)
