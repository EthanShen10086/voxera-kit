package analytics

import "time"

// Event represents a single user or system action tracked for analytics.
type Event struct {
	// ID is a globally unique identifier for this event.
	ID string
	// Name describes the action, e.g. "page_view", "button_click", "purchase".
	Name string
	// UserID identifies the user who triggered the event.
	UserID string
	// SessionID groups events within a single user session.
	SessionID string
	// TenantID scopes the event to a specific tenant in multi-tenant setups.
	TenantID string
	// Timestamp records when the event occurred.
	Timestamp time.Time
	// Properties holds arbitrary key-value metadata attached to the event.
	Properties map[string]any
	// Ctx holds environmental metadata auto-captured with the event.
	// The field is named Ctx to avoid colliding with context.Context parameters.
	Ctx EventContext
}

// EventContext holds environmental metadata auto-captured with each event.
type EventContext struct {
	// Platform identifies the client platform: "web", "ios", "android", or "desktop".
	Platform string
	// AppVersion is the semantic version of the application.
	AppVersion string
	// Locale is the user's locale string, e.g. "en-US".
	Locale string
	// UserAgent is the raw User-Agent header from the client.
	UserAgent string
	// IP is the client IP address.
	IP string
	// PageURL is the full URL of the page where the event was triggered.
	PageURL string
	// Referrer is the HTTP Referer header value.
	Referrer string
	// UTMSource is the utm_source marketing parameter.
	UTMSource string
	// UTMMedium is the utm_medium marketing parameter.
	UTMMedium string
	// UTMCampaign is the utm_campaign marketing parameter.
	UTMCampaign string
	// UTMTerm is the utm_term marketing parameter.
	UTMTerm string
	// UTMContent is the utm_content marketing parameter.
	UTMContent string
	// DeviceID is a persistent identifier for the physical device.
	DeviceID string
	// ScreenWidth is the device screen width in pixels.
	ScreenWidth int
	// ScreenHeight is the device screen height in pixels.
	ScreenHeight int
}

// EventQuery specifies filters for querying raw events.
type EventQuery struct {
	// UserID limits results to a specific user.
	UserID string
	// TenantID limits results to a specific tenant.
	TenantID string
	// Names filters results to events matching any of the listed names.
	Names []string
	// From is the inclusive lower bound of the time range.
	From time.Time
	// To is the exclusive upper bound of the time range.
	To time.Time
	// Limit caps the number of returned events.
	Limit int
	// Offset skips the first N matching events (for pagination).
	Offset int
	// OrderBy controls sort order: "timestamp_asc" or "timestamp_desc".
	OrderBy string
}
