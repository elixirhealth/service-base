package cmd

import (
	"os"
	"testing"

	"github.com/elxirhealth/service-base/version"
)

func TestWriteBanner(t *testing.T) {
	// just test that there are no panics from mismatched template vars
	writeBanner(os.Stdout, "Test", version.Current)
}
