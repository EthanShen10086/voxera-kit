// Package scraper provides a pluggable web scraping engine abstraction
// supporting proxy rotation, rate limiting, and anti-ban strategies.
package scraper

import (
	"context"
	"net/http"
	"time"
)

// Post represents a unified social media post model across platforms.
type Post struct {
	ID          string
	PlatformID  string
	Platform    string
	AuthorID    string
	AuthorName  string
	Content     string
	MediaURLs   []string
	PublishedAt time.Time
	FetchedAt   time.Time
	URL         string
	Metadata    map[string]any
}

// FetchRequest represents a request to be made by the Fetcher.
type FetchRequest struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    []byte
	Timeout time.Duration
}

// FetchResponse represents the response from a fetch operation.
type FetchResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Duration   time.Duration
}

// Fetcher is the HTTP request abstraction supporting proxy rotation and retries.
type Fetcher interface {
	Fetch(ctx context.Context, req *FetchRequest) (*FetchResponse, error)
}

// Parser extracts structured Post data from raw fetch responses.
type Parser interface {
	Parse(ctx context.Context, resp *FetchResponse) ([]*Post, error)
}

// Source defines a data source configuration for tracking a platform.
type Source struct {
	Platform     string
	PollInterval time.Duration
	Priority     int
	Config       map[string]any
}

// Proxy represents a proxy server configuration.
type Proxy struct {
	URL      string
	Protocol string
	Username string
	Password string
	Healthy  bool
}

// ProxyPool manages a pool of proxy servers with rotation and health checking.
type ProxyPool interface {
	Next(ctx context.Context) (*Proxy, error)
	Add(proxy *Proxy) error
	Remove(url string) error
	HealthCheck(ctx context.Context) error
	Size() int
}

// RateLimitPolicy defines rate limiting behavior for a specific platform.
type RateLimitPolicy struct {
	Platform       string
	RequestsPerMin int
	BurstSize      int
	BackoffBase    time.Duration
	BackoffMax     time.Duration
	JitterPercent  int
}

// RateLimiter enforces rate limiting for scraping requests.
type RateLimiter interface {
	Wait(ctx context.Context, platform string) error
	RecordSuccess(platform string)
	RecordFailure(platform string, statusCode int)
}

// Tracker orchestrates periodic fetching and new post detection for tracked accounts.
type Tracker interface {
	Track(ctx context.Context, source *Source, accountID string) error
	Untrack(ctx context.Context, platform string, accountID string) error
	OnNewPost(handler func(post *Post))
}
