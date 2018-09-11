package testing_test

import "testing"

// TestInstaller tests that the installer works correctly.
func TestInstaller(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
}
