package model

import "context"

type BulkResult struct {
	Accepted   int
	MessageIDs []string
	RawStatus  string
}

type Provider interface {
	SendBulk(ctx context.Context, sender string, receptors []string, message string) (BulkResult, error)
	SendLookup(ctx context.Context, receptor, template string, tokens map[string]string) error
}
