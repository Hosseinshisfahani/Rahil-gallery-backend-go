package sms

import "errors"

var (
	ErrInvalidReceptor = errors.New("invalid sms receptor")
	ErrProviderFailed  = errors.New("sms provider failed")
	ErrDisabled        = errors.New("sms provider disabled")
)
