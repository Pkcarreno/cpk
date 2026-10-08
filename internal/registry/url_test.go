package registry_test

import (
	"testing"

	"github.com/pkcarreno/cpk/internal/registry"
)

func TestResolveRegistryURL(t *testing.T) {
	t.Run("returns default URL when custom argument and environment variable are empty", func(t *testing.T) {
		actualURL := registry.ResolveRegistryURL("")

		expectedURL := "https://registry.npmjs.org"

		if actualURL != expectedURL {
			t.Fatalf("expected: %q, got: %q", expectedURL, actualURL)
		}
	})

	t.Run("returns environment variable when custom argument is empty", func(t *testing.T) {
		environmentURL := "https://private.registry.org"
		t.Setenv("NPM_CONFIG_REGISTRY", environmentURL)

		actualURL := registry.ResolveRegistryURL("")

		if actualURL != environmentURL {
			t.Fatalf("expected: %q, got: %q", environmentURL, actualURL)
		}
	})

	t.Run("returns custom argument when environment variable is also set", func(t *testing.T) {
		environmentURL := "https://private.registry.org"
		t.Setenv("NPM_CONFIG_REGISTRY", environmentURL)

		customURL := "https://custom.registry.org"

		actualURL := registry.ResolveRegistryURL(customURL)

		if actualURL != customURL {
			t.Fatalf("expected: %q, got: %q", customURL, actualURL)
		}
	})
}
