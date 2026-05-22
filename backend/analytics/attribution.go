package analytics

// Attribution holds UTM / channel attribution data for a user.
type Attribution struct {
	// FirstTouch is the attribution data from the user's first interaction.
	FirstTouch TouchPoint
	// LastTouch is the attribution data from the user's most recent interaction.
	LastTouch TouchPoint
}

// TouchPoint represents a single marketing attribution touch.
type TouchPoint struct {
	// Source is the utm_source value.
	Source string
	// Medium is the utm_medium value.
	Medium string
	// Campaign is the utm_campaign value.
	Campaign string
	// Term is the utm_term value.
	Term string
	// Content is the utm_content value.
	Content string
	// Referrer is the HTTP referrer at the time of this touch.
	Referrer string
}
