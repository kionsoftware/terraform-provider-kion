package crud

import (
	"os"
	"testing"
)

// Fixture version: the golden output was rendered against v3_16.
func TestMain(m *testing.M) {
	sdkVersion = "v3_16"
	os.Exit(m.Run())
}
