package sms

import (
	"log"

	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	domainsms "github.com/rahil-gallery/rahil-gallery-server/internal/domain/sms"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/sms/kavenegar"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/sms/noop"
)

// NewProvider returns Noop when disabled; otherwise the Kavenegar HTTP client.
func NewProvider(cfg config.Config) domainsms.Provider {
	if !cfg.KavenegarEnabled {
		log.Printf("sms: provider=noop (KAVENEGAR_ENABLED=false)")
		return noop.New()
	}
	if cfg.KavenegarAPIKey == "" {
		log.Printf("sms: provider=noop (KAVENEGAR_ENABLED=true but API key empty)")
		return noop.New()
	}
	log.Printf("sms: provider=kavenegar template=%s sender=%s", cfg.KavenegarBirthdayTemplate, cfg.KavenegarSender)
	return kavenegar.New(cfg.KavenegarAPIKey)
}
