package model

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidReceptor = errors.New("invalid sms receptor")
	ErrProviderFailed  = errors.New("sms provider failed")
)
