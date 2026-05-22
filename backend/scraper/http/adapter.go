// Package http provides a standard library based Fetcher implementation
// with proxy rotation, User-Agent pooling, and retry logic.
package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/EthanShen10086/voxera-kit/scraper"
)

// Config holds configuration for the HTTP fetcher.
type Config struct {
	DefaultTimeout time.Duration
	MaxRetries     int
	UserAgents     []string
	ProxyPool      scraper.ProxyPool
}

// Adapter implements Fetcher using the Go standard library HTTP client.
type Adapter struct {
	config Config
	client *http.Client
}

// New creates a new HTTP fetcher adapter.
func New(cfg Config) *Adapter {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = 30 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if len(cfg.UserAgents) == 0 {
		cfg.UserAgents = defaultUserAgents()
	}
	return &Adapter{config: cfg, client: &http.Client{
		Timeout: cfg.DefaultTimeout,
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{MinVersion: tls.VersionTLS12},
			DisableKeepAlives: false,
			MaxIdleConns:      100,
			IdleConnTimeout:   90 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}}
}

// Fetch executes an HTTP request with proxy rotation and UA randomization.
func (a *Adapter) Fetch(ctx context.Context, req *scraper.FetchRequest) (*scraper.FetchResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("scraper/http: create request: %w", err)
	}

	httpReq.Header.Set("User-Agent", a.randomUA())
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	if a.config.ProxyPool != nil {
		proxy, proxyErr := a.config.ProxyPool.Next(ctx)
		if proxyErr == nil && proxy != nil {
			proxyURL, _ := url.Parse(proxy.URL)
			a.client.Transport.(*http.Transport).Proxy = http.ProxyURL(proxyURL)
		}
	}

	start := time.Now()
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("scraper/http: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if readErr != nil {
			break
		}
	}

	return &scraper.FetchResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
		Duration:   time.Since(start),
	}, nil
}

func (a *Adapter) randomUA() string {
	return a.config.UserAgents[rand.IntN(len(a.config.UserAgents))]
}

func defaultUserAgents() []string {
	return []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:126.0) Gecko/20100101 Firefox/126.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:126.0) Gecko/20100101 Firefox/126.0",
	}
}
