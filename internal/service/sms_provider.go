package service

import (
	"context"
	"log"

	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/model"
)

// NewProvider returns Noop when disabled; otherwise the Kavenegar HTTP client.
func NewProvider(cfg config.Config) model.Provider {
	if !cfg.KavenegarEnabled {
		log.Printf("sms: provider=noop (KAVENEGAR_ENABLED=false)")
		return NewNoop()
	}
	if cfg.KavenegarAPIKey == "" {
		log.Printf("sms: provider=noop (KAVENEGAR_ENABLED=true but API key empty)")
		return NewNoop()
	}
	log.Printf("sms: provider=kavenegar template=%s sender=%s", cfg.KavenegarBirthdayTemplate, cfg.KavenegarSender)
	return NewKavenegar(cfg.KavenegarAPIKey)
}

// NoopProvider logs intent and succeeds. Used when KAVENEGAR_ENABLED=false.
type NoopProvider struct{}

func NewNoop() *NoopProvider { return &NoopProvider{} }

func (p *NoopProvider) SendBulk(ctx context.Context, sender string, receptors []string, message string) (model.BulkResult, error) {
	log.Printf("sms noop SendBulk sender=%s receptors=%d message_len=%d", sender, len(receptors), len(message))
	return model.BulkResult{Accepted: len(receptors), RawStatus: "noop"}, nil
}

func (p *NoopProvider) SendLookup(ctx context.Context, receptor, template string, tokens map[string]string) error {
	log.Printf("sms noop SendLookup receptor=%s template=%s tokens=%v", receptor, template, tokens)
	return nil
}

var _ model.Provider = (*NoopProvider)(nil)
