package version_test

import (
	"testing"

	"github.com/pkcarreno/cpk/internal/version"
)

func TestCurrent(t *testing.T) {
	t.Run("returns default dev version when unpopulated", func(t *testing.T) {
		buildMetadata := version.Current()

		if buildMetadata.Version != "dev" {
			t.Fatalf("expected version dev, got %q", buildMetadata.Version)
		}
	})
}
