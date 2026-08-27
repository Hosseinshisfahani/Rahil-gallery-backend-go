package sms

import (
	"strings"
	"unicode"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
)

// NormalizeReceptor turns a stored CRM phone into Kavenegar-friendly 98xxxxxxxxxx.
// Returns ("", ErrInvalidReceptor) if it is not a plausible Iranian mobile.
func NormalizeReceptor(raw string) (string, error) {
	s := customer.NormalizeDigits(strings.TrimSpace(raw))
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	digits := b.String()
	switch {
	case strings.HasPrefix(digits, "98") && len(digits) == 12:
		digits = digits[2:]
	case strings.HasPrefix(digits, "0") && len(digits) == 11:
		digits = digits[1:]
	}
	// National mobile: 9xxxxxxxxx (10 digits)
	if len(digits) != 10 || digits[0] != '9' {
		return "", ErrInvalidReceptor
	}
	return "98" + digits, nil
}
