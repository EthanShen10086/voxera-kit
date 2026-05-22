// Package regex provides a regexp-based PII redactor with pre-built rules for
// common patterns such as emails, phone numbers, and credit cards.
package regex

import (
	"fmt"
	"regexp"

	"github.com/EthanShen10086/voxera-kit/pii"
)

var (
	emailPattern      = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	phonePattern      = regexp.MustCompile(`\b(\+?1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b`)
	creditCardPattern = regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`)
	ssnPattern        = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	ipPattern         = regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)
)

// Redactor applies PII redaction rules using regular expressions.
type Redactor struct {
	cfg pii.Config
}

// NewRedactor creates a new Redactor with the given configuration.
func NewRedactor(cfg pii.Config) *Redactor {
	return &Redactor{cfg: cfg}
}

// DefaultRules returns a set of pre-built rules for common PII patterns
// including email, phone, credit card, SSN, and IP address.
func DefaultRules() []pii.Rule {
	return []pii.Rule{
		{Pattern: emailPattern, Replacement: "[EMAIL REDACTED]"},
		{Pattern: phonePattern, Replacement: "[PHONE REDACTED]"},
		{Pattern: creditCardPattern, Replacement: "[CREDIT CARD REDACTED]"},
		{Pattern: ssnPattern, Replacement: "[SSN REDACTED]"},
		{Pattern: ipPattern, Replacement: "[IP REDACTED]"},
	}
}

// Redact applies all configured rules sequentially to the given string value.
func (r *Redactor) Redact(value string) string {
	result := value
	for _, rule := range r.cfg.Rules {
		if rule.Pattern != nil {
			replacement := rule.Replacement
			if replacement == "" {
				replacement = r.cfg.DefaultMask
			}
			result = rule.Pattern.ReplaceAllString(result, replacement)
		}
	}
	return result
}

// RedactFields recursively processes map values, applying redaction rules to
// string values and descending into nested maps.
func (r *Redactor) RedactFields(data map[string]any) map[string]any {
	result := make(map[string]any, len(data))
	for k, v := range data {
		result[k] = r.redactValue(k, v)
	}
	return result
}

func (r *Redactor) redactValue(fieldName string, value any) any {
	for _, rule := range r.cfg.Rules {
		if rule.FieldName != "" && rule.FieldName == fieldName {
			mask := rule.Replacement
			if mask == "" {
				mask = r.cfg.DefaultMask
			}
			return mask
		}
	}

	switch v := value.(type) {
	case string:
		return r.Redact(v)
	case map[string]any:
		return r.RedactFields(v)
	case fmt.Stringer:
		return r.Redact(v.String())
	default:
		return value
	}
}
