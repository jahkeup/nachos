//go:build darwin
// +build darwin

package nachos

import (
	"runtime"
	"testing"
)

// TestAssertDarwin verifies the test run on darwin.
func TestAssertDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Fatal("must only run on darwin")
	}
}
