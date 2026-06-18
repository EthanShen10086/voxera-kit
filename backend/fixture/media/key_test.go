package media

import "testing"

func TestObjectKey(t *testing.T) {
	got := ObjectKey("users", "u1", "avatar.png")
	want := "users/u1/avatar.png"
	if got != want {
		t.Fatalf("ObjectKey() = %q, want %q", got, want)
	}
}

func TestAudioObjectKey(t *testing.T) {
	got := AudioObjectKey("user-1", "rec-1")
	want := "users/user-1/audio/rec-1.wav"
	if got != want {
		t.Fatalf("AudioObjectKey() = %q, want %q", got, want)
	}
}

func TestSampleBytesNonEmpty(t *testing.T) {
	for name, sample := range map[string][]byte{
		"text": SampleText(),
		"wav":  SampleWAV(),
		"png":  SamplePNG(),
	} {
		if len(sample) == 0 {
			t.Fatalf("%s sample is empty", name)
		}
	}
}
