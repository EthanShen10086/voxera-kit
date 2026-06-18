package fixture

import "time"

// FixedTime returns a stable UTC timestamp for deterministic tests.
func FixedTime() time.Time {
	return time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
}

// MustParseRFC3339 parses an RFC3339 timestamp and panics on failure.
func MustParseRFC3339(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic("fixture: parse RFC3339: " + err.Error())
	}
	return t.UTC()
}

// TruncateUTC truncates t to whole seconds in UTC.
func TruncateUTC(t time.Time) time.Time {
	return t.UTC().Truncate(time.Second)
}
