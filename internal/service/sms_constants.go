package service

import "time"

const (
	LookupFallbackToken = "مشتری"

	BirthdayStatusSent                = "sent"
	BirthdayStatusFailed              = "failed"
	BirthdayStatusSkippedInvalidPhone = "skipped_invalid_phone"

	MaxBulkMessageRunes = 900
	MaxBulkRecipients    = 20000
	MaxSellerNoteRunes   = 4000
	MaxSignatureBytes    = 2 << 20 // 2MB
	BirthdayRunTimeout   = 10 * time.Minute
)
