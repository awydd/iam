package code

import "net/http"

type Code int

const (
	Success Code = 0 + iota
	Fail
)

const (
	BadRequest Code = 1000 + iota
	Unauthorized
	Forbidden
	NotFound
	Conflict
	TooManyRequests
	InternalError
)

type meta struct {
	message    string
	httpStatus int
}

var registry = map[Code]meta{
	Success: {"success", http.StatusOK},
	Fail:    {"fail", http.StatusBadRequest},

	BadRequest:      {"bad request", http.StatusBadRequest},
	Unauthorized:    {"unauthenticated", http.StatusUnauthorized},
	Forbidden:       {"forbidden", http.StatusForbidden},
	NotFound:        {"resource not found", http.StatusNotFound},
	Conflict:        {"resource conflict", http.StatusConflict},
	TooManyRequests: {"too many requests", http.StatusTooManyRequests},
	InternalError:   {"internal server error", http.StatusInternalServerError},
}

func (c Code) Message() string {
	if m, ok := registry[c]; ok {
		return m.message
	}
	return "unknown error"
}

func (c Code) HTTPStatus() int {
	if m, ok := registry[c]; ok {
		return m.httpStatus
	}
	return http.StatusInternalServerError
}

func (c Code) Int() int { return int(c) }
