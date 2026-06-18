package containers

import "testing"

func TestSanitizeBucketName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "voxera-test"},
		{"voxera-test", "voxera-test"},
		{"TestMinIOObjectStoreContract", "testminioobjectstorecontract"},
		{"vm-Test/Foo Bar", "vm-test-foo-bar"},
	}
	for _, tc := range tests {
		if got := sanitizeBucketName(tc.in); got != tc.want {
			t.Fatalf("sanitizeBucketName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
