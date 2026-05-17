// Package memory provides an in-process implementation of the security.IPFilter interface.
package memory

import (
	"sync"

	"github.com/EthanShen10086/voxera-kit/security"
)

// Adapter implements security.IPFilter using in-memory maps protected by a read-write mutex.
type Adapter struct {
	mu        sync.RWMutex
	mode      security.FilterMode
	whitelist map[string]bool
	blacklist map[string]bool
}

// New creates a new in-memory IP filter from the given configuration.
func New(cfg security.IPFilterConfig) *Adapter {
	a := &Adapter{
		mode:      cfg.Mode,
		whitelist: make(map[string]bool, len(cfg.WhitelistIPs)),
		blacklist: make(map[string]bool, len(cfg.BlacklistIPs)),
	}
	for _, ip := range cfg.WhitelistIPs {
		a.whitelist[ip] = true
	}
	for _, ip := range cfg.BlacklistIPs {
		a.blacklist[ip] = true
	}
	return a
}

// IsAllowed reports whether the given IP address is permitted under the current filter mode.
func (a *Adapter) IsAllowed(ip string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	switch a.mode {
	case security.Whitelist:
		return a.whitelist[ip]
	case security.Blacklist:
		return !a.blacklist[ip]
	case security.Both:
		if a.whitelist[ip] {
			return true
		}
		return !a.blacklist[ip]
	default:
		return false
	}
}

// AddToWhitelist adds IP addresses to the allow list.
func (a *Adapter) AddToWhitelist(ips ...string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, ip := range ips {
		a.whitelist[ip] = true
	}
}

// AddToBlacklist adds IP addresses to the deny list.
func (a *Adapter) AddToBlacklist(ips ...string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, ip := range ips {
		a.blacklist[ip] = true
	}
}

// RemoveFromWhitelist removes IP addresses from the allow list.
func (a *Adapter) RemoveFromWhitelist(ips ...string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, ip := range ips {
		delete(a.whitelist, ip)
	}
}

// RemoveFromBlacklist removes IP addresses from the deny list.
func (a *Adapter) RemoveFromBlacklist(ips ...string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, ip := range ips {
		delete(a.blacklist, ip)
	}
}

// Reload replaces the current configuration atomically.
func (a *Adapter) Reload(cfg security.IPFilterConfig) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.mode = cfg.Mode
	a.whitelist = make(map[string]bool, len(cfg.WhitelistIPs))
	for _, ip := range cfg.WhitelistIPs {
		a.whitelist[ip] = true
	}
	a.blacklist = make(map[string]bool, len(cfg.BlacklistIPs))
	for _, ip := range cfg.BlacklistIPs {
		a.blacklist[ip] = true
	}
}
