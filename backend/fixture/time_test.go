package fixture

import (
	"testing"
	"time"
)

func TestFixedTime(t *testing.T) {
	got := FixedTime()
	want := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("FixedTime() = %v, want %v", got, want)
	}
}

func TestMustParseRFC3339(t *testing.T) {
	got := MustParseRFC3339("2024-06-01T08:30:00Z")
	want := time.Date(2024, 6, 1, 8, 30, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("MustParseRFC3339() = %v, want %v", got, want)
	}
}

func TestTruncateUTC(t *testing.T) {
	input := time.Date(2024, 6, 1, 8, 30, 1, 500, time.FixedZone("CET", 3600))
	got := TruncateUTC(input)
	want := time.Date(2024, 6, 1, 7, 30, 1, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("TruncateUTC() = %v, want %v", got, want)
	}
}
