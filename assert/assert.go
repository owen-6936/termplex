package assert

import (
	"strings"
	"testing"
)

// Contains checks if a string `s` contains the substring `substr`.
// If it doesn't, it fails the test with a descriptive message.
func Contains(t *testing.T, s, substr string) {
	t.Helper() // Marks this function as a test helper.
	if !strings.Contains(s, substr) {
		t.Errorf("expected string to contain %q, but it did not. Full string: %q", substr, s)
	}
}

// NoError checks if an error is nil.
// If it's not, it fails the test immediately.
func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
}

// True checks if a boolean condition is true.
func True(t *testing.T, condition bool, msg string, args ...interface{}) {
	t.Helper()
	if !condition {
		t.Errorf(msg, args...)
	}
}
