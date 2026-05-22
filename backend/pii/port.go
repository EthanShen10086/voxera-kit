// Package pii provides tools for detecting and redacting personally
// identifiable information from strings and structured data.
package pii

import "regexp"

// Redactor defines the interface for PII redaction operations.
type Redactor interface {
	// Redact applies all configured rules to a single string value.
	Redact(value string) string
	// RedactFields recursively processes a map, redacting values that match
	// configured rules.
	RedactFields(data map[string]any) map[string]any
}

// Rule defines a single PII detection rule that matches either by exact field
// name or by a regex pattern applied to the content.
type Rule struct {
	FieldName   string
	Pattern     *regexp.Regexp
	Replacement string
}

// Config holds the configuration for a Redactor instance.
type Config struct {
	Rules       []Rule
	DefaultMask string
}
