package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// DefaultDevJWTSecret is the shared fallback for local/docker dev when JWT_ACCESS_SECRET is unset.
// Keep in sync with docker-compose.yml and scripts/run-api-dev.sh.
const DefaultDevJWTSecret = "dev-docker-secret-change-in-production"

const minJWTSecretLength = 32

type Config struct {
	Env                   string
	Host                  string
	Port                  int
	DatabaseURL           string
	JWTAccessSecret       string
	JWTAccessTTL          time.Duration
	JWTRefreshTTL         time.Duration
	CustomerSignaturesDir string
	// Kavenegar / SMS
	KavenegarAPIKey           string
	KavenegarSender           string
	KavenegarBirthdayTemplate string
	KavenegarEnabled          bool
	SMSBirthdayCron           string
	SMSBirthdayTZ             string
	SMSBulkBatchSize          int
	SMSBulkMaxConcurrency     int
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: loading .env: %v", err)
	}

	port, err := strconv.Atoi(getEnv("APP_PORT", "8081"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid APP_PORT: %w", err)
	}

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	secret := strings.TrimSpace(os.Getenv("JWT_ACCESS_SECRET"))
	usingDefaultSecret := secret == ""
	if usingDefaultSecret {
		secret = DefaultDevJWTSecret
	}

	bulkBatchSize, err := strconv.Atoi(getEnv("SMS_BULK_BATCH_SIZE", "200"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid SMS_BULK_BATCH_SIZE: %w", err)
	}
	bulkMaxConcurrency, err := strconv.Atoi(getEnv("SMS_BULK_MAX_CONCURRENCY", "3"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid SMS_BULK_MAX_CONCURRENCY: %w", err)
	}

	cfg := Config{
		Env:                       getEnv("APP_ENV", "development"),
		Host:                      getEnv("APP_HOST", "0.0.0.0"),
		Port:                      port,
		DatabaseURL:               os.Getenv("DATABASE_URL"),
		JWTAccessSecret:           secret,
		JWTAccessTTL:              accessTTL,
		JWTRefreshTTL:             refreshTTL,
		CustomerSignaturesDir:     getEnv("CUSTOMER_SIGNATURES_DIR", "data/customer-signatures"),
		KavenegarAPIKey:           strings.TrimSpace(os.Getenv("KAVENEGAR_API_KEY")),
		KavenegarSender:           getEnv("KAVENEGAR_SENDER", getEnv("SMS_SENDER_ID", "")),
		KavenegarBirthdayTemplate: strings.TrimSpace(getEnv("KAVENEGAR_BIRTHDAY_TEMPLATE", "birthday")),
		KavenegarEnabled:          getEnv("KAVENEGAR_ENABLED", "false") == "true",
		SMSBirthdayCron:           getEnv("SMS_BIRTHDAY_CRON", "0 9 * * *"),
		SMSBirthdayTZ:             getEnv("SMS_BIRTHDAY_TZ", "Asia/Tehran"),
		SMSBulkBatchSize:          bulkBatchSize,
		SMSBulkMaxConcurrency:     bulkMaxConcurrency,
	}

	if err := cfg.Validate(usingDefaultSecret); err != nil {
		return Config{}, err
	}

	if usingDefaultSecret {
		log.Printf("warning: JWT_ACCESS_SECRET not set; using development default (fingerprint=%s)", jwtSecretFingerprint(secret))
	} else {
		log.Printf("auth: JWT secret configured (fingerprint=%s)", jwtSecretFingerprint(secret))
	}

	return cfg, nil
}

func (c Config) Validate(usingDefaultSecret bool) error {
	if c.JWTAccessTTL < time.Minute {
		return fmt.Errorf("JWT_ACCESS_TTL must be at least 1m")
	}

	if c.JWTRefreshTTL < c.JWTAccessTTL {
		return fmt.Errorf("JWT_REFRESH_TTL must be greater than JWT_ACCESS_TTL")
	}

	isProduction := strings.EqualFold(c.Env, "production") || strings.EqualFold(c.Env, "prod")
	if isProduction {
		if usingDefaultSecret {
			return fmt.Errorf("JWT_ACCESS_SECRET must be set in production")
		}
		if len(c.JWTAccessSecret) < minJWTSecretLength {
			return fmt.Errorf("JWT_ACCESS_SECRET must be at least %d characters in production", minJWTSecretLength)
		}
	} else if usingDefaultSecret && len(c.JWTAccessSecret) < minJWTSecretLength {
		return fmt.Errorf("development JWT default is too short; set JWT_ACCESS_SECRET in .env")
	}

	if c.KavenegarEnabled {
		if strings.TrimSpace(c.KavenegarAPIKey) == "" {
			return fmt.Errorf("KAVENEGAR_API_KEY is required when KAVENEGAR_ENABLED=true")
		}
		if strings.TrimSpace(c.KavenegarBirthdayTemplate) == "" {
			return fmt.Errorf("KAVENEGAR_BIRTHDAY_TEMPLATE is required when KAVENEGAR_ENABLED=true")
		}
		if c.SMSBulkBatchSize < 1 || c.SMSBulkBatchSize > 200 {
			return fmt.Errorf("SMS_BULK_BATCH_SIZE must be between 1 and 200")
		}
		if c.SMSBulkMaxConcurrency < 1 {
			return fmt.Errorf("SMS_BULK_MAX_CONCURRENCY must be at least 1")
		}
	}

	return nil
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func jwtSecretFingerprint(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:4])
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
