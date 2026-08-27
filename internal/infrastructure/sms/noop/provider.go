package noop

import (
	"context"
	"log"

	domainsms "github.com/rahil-gallery/rahil-gallery-server/internal/domain/sms"
)

// Provider logs intent and succeeds. Used when KAVENEGAR_ENABLED=false.
type Provider struct{}

func New() *Provider { return &Provider{} }

func (p *Provider) SendBulk(ctx context.Context, sender string, receptors []string, message string) (domainsms.BulkResult, error) {
	log.Printf("sms noop SendBulk sender=%s receptors=%d message_len=%d", sender, len(receptors), len(message))
	return domainsms.BulkResult{Accepted: len(receptors), RawStatus: "noop"}, nil
}

func (p *Provider) SendLookup(ctx context.Context, receptor, template string, tokens map[string]string) error {
	log.Printf("sms noop SendLookup receptor=%s template=%s tokens=%v", receptor, template, tokens)
	return nil
}

// Compile-time check
var _ domainsms.Provider = (*Provider)(nil)
