package consts

import "time"

const SessionTTL = 8 * time.Hour

const (
	CookieAccessToken  = "cat"
	CookieRefreshToken = "crt"
	CookieSessionID    = "csid"
)
