// Package memory provides an in-memory implementation of the scraper.ProxyPool.
package memory

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/EthanShen10086/voxera-kit/scraper"
)

// ProxyPool is an in-memory proxy pool with round-robin rotation.
type ProxyPool struct {
	mu      sync.RWMutex
	proxies []*scraper.Proxy
	index   atomic.Int64
}

// NewProxyPool creates a new in-memory proxy pool.
func NewProxyPool() *ProxyPool {
	return &ProxyPool{proxies: make([]*scraper.Proxy, 0)}
}

// Next returns the next available proxy using round-robin.
func (p *ProxyPool) Next(_ context.Context) (*scraper.Proxy, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.proxies) == 0 {
		return nil, fmt.Errorf("scraper/memory: proxy pool is empty")
	}
	idx := p.index.Add(1) % int64(len(p.proxies))
	return p.proxies[idx], nil
}

// Add adds a proxy to the pool.
func (p *ProxyPool) Add(proxy *scraper.Proxy) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	proxy.Healthy = true
	p.proxies = append(p.proxies, proxy)
	return nil
}

// Remove removes a proxy by URL.
func (p *ProxyPool) Remove(proxyURL string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, px := range p.proxies {
		if px.URL == proxyURL {
			p.proxies = append(p.proxies[:i], p.proxies[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("scraper/memory: proxy %q not found", proxyURL)
}

// HealthCheck marks unhealthy proxies (stub implementation).
func (p *ProxyPool) HealthCheck(_ context.Context) error {
	return nil
}

// Size returns the number of proxies in the pool.
func (p *ProxyPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.proxies)
}
