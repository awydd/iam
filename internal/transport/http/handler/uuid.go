package handler

import (
	"errors"

	"github.com/google/uuid"
)

type UUIDParam uuid.UUID

func (u *UUIDParam) UnmarshalParam(s string) error {
	if s == "" {
		return errors.New("session_id is required")
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return err
	}
	*u = UUIDParam(parsed)
	return nil
}
