// Package security defines the port interfaces for IP filtering and hotlink protection.
// It provides abstractions for whitelist/blacklist-based IP access control and
// signed URL generation for resource protection.
package security

import "time"

// FilterMode determines which IP lists are evaluated.
type FilterMode int

const (
	// Whitelist only allows IPs explicitly listed.
	Whitelist FilterMode = iota
	// Blacklist only denies IPs explicitly listed.
	Blacklist
	// Both evaluates whitelist first, then blacklist.
	Both
)

// IPFilterConfig holds the parameters for constructing an IP filter.
type IPFilterConfig struct {
	// Enabled controls whether IP filtering is active.
	Enabled bool
	// Mode selects the filtering strategy.
	Mode FilterMode
	// WhitelistIPs is the set of allowed IP addresses.
	WhitelistIPs []string
	// BlacklistIPs is the set of denied IP addresses.
	BlacklistIPs []string
}

// IPFilter manages IP-based access control.
// Implementations must be safe for concurrent use.
type IPFilter interface {
	// IsAllowed reports whether the given IP address is permitted.
	IsAllowed(ip string) bool
	// AddToWhitelist adds IP addresses to the allow list.
	AddToWhitelist(ips ...string)
	// AddToBlacklist adds IP addresses to the deny list.
	AddToBlacklist(ips ...string)
	// RemoveFromWhitelist removes IP addresses from the allow list.
	RemoveFromWhitelist(ips ...string)
	// RemoveFromBlacklist removes IP addresses from the deny list.
	RemoveFromBlacklist(ips ...string)
	// Reload replaces the current configuration atomically.
	Reload(config IPFilterConfig)
}

// HotlinkConfig holds the parameters for hotlink protection.
type HotlinkConfig struct {
	// Enabled controls whether hotlink protection is active.
	Enabled bool
	// AllowedDomains is the set of domains permitted to link to resources.
	AllowedDomains []string
	// SigningSecret is the HMAC key used for URL signing.
	SigningSecret string
}

// HotlinkProtection provides referer validation and signed URL management.
type HotlinkProtection interface {
	// IsValidReferer reports whether the referer header matches an allowed domain.
	IsValidReferer(referer string, allowedDomains []string) bool
	// GenerateSignedURL produces a URL with an HMAC signature and expiry timestamp.
	GenerateSignedURL(rawURL string, expiry time.Duration) (string, error)
	// ValidateSignedURL verifies the signature and expiry of a signed URL,
	// returning the original URL if valid.
	ValidateSignedURL(signedURL string) (string, error)
}
