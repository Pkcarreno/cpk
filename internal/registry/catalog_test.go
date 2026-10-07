package registry_test

import (
	"errors"
	"testing"

	"github.com/pkcarreno/cpk/internal/registry"
)

func TestValidatePackageManager(t *testing.T) {
	testCases := []struct {
		packageManager string
	}{
		{packageManager: "npm"},
		{packageManager: "yarn"},
		{packageManager: "pnpm"},
	}

	for _, testCase := range testCases {
		description := "returns nil when the package manager is " + testCase.packageManager

		t.Run(description, func(t *testing.T) {
			err := registry.ValidatePackageManager(testCase.packageManager)

			if err != nil {
				t.Fatalf("expected nil, got: %v", err)
			}
		})
	}

	t.Run("returns ErrUnsupportedPackageManager when the package manager is uppercase", func(t *testing.T) {
		err := registry.ValidatePackageManager("PNPM")

		if !errors.Is(err, registry.ErrUnsupportedPackageManager) {
			t.Fatalf("expected ErrUnsupportedPackageManager, got: %v", err)
		}
	})

	t.Run("returns ErrUnsupportedPackageManager when the package manager is empty", func(t *testing.T) {
		err := registry.ValidatePackageManager("")

		if !errors.Is(err, registry.ErrUnsupportedPackageManager) {
			t.Fatalf("expected ErrUnsupportedPackageManager, got: %v", err)
		}
	})

	t.Run("returns ErrUnsupportedPackageManager when the package manager is not supported", func(t *testing.T) {
		err := registry.ValidatePackageManager("bun")

		if !errors.Is(err, registry.ErrUnsupportedPackageManager) {
			t.Fatalf("expected ErrUnsupportedPackageManager, got: %v", err)
		}
	})
}
