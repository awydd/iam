package biz

import (
	"errors"
)

// user
var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountNotActive   = errors.New("account is not active")
)
