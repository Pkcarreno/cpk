package manifest_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkcarreno/cpk/internal/manifest"
)

func TestReadNotFound(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "package.json")

	_, err := manifest.Read(missingPath)
	if err == nil {
		t.Fatalf("expected an error, but got nil")
	}
	if !errors.Is(err, manifest.ErrManifestNotFound) {
		t.Fatalf("expected ErrManifestNotFound, got: %v", err)
	}
	expectedErrorMessage := "the package.json file does not exist"
	if err.Error() != expectedErrorMessage {
		t.Fatalf("expected %q, got %q", expectedErrorMessage, err.Error())
	}
}

func TestParsePackageSpec(t *testing.T) {
	t.Run("parses valid spec with hash", func(t *testing.T) {
		spec, err := manifest.ParsePackageSpec("pnpm@9.1.0+sha512.abcdef")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if spec.Name != "pnpm" || spec.Version != "9.1.0" || spec.Hash != "+sha512.abcdef" {
			t.Fatalf("unexpected spec: %v", spec)
		}
	})

	t.Run("parses valid spec without hash", func(t *testing.T) {
		spec, err := manifest.ParsePackageSpec("yarn@4.1.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if spec.Name != "yarn" || spec.Version != "4.1.0" || spec.Hash != "" {
			t.Fatalf("unexpected spec: %+v", spec)
		}
	})

	t.Run("returns error on invalid spec", func(t *testing.T) {
		_, err := manifest.ParsePackageSpec("invalid-spec")
		if !errors.Is(err, manifest.ErrInvalidPackageSpec) {
			t.Fatalf("expected ErrInvalidPackageSpec, got: %v", err)
		}
	})
}

func TestReadPackageSpec(t *testing.T) {
	t.Run("returns ErrManifestNotFound when file does not exist", func(t *testing.T) {
		missingPath := filepath.Join(t.TempDir(), "package.json")
		_, err := manifest.ReadPackageSpec(missingPath)
		if !errors.Is(err, manifest.ErrManifestNotFound) {
			t.Fatalf("expected ErrManifestNotFound, got: %v", err)
		}
	})

	t.Run("return ErrMissingPackageManager when key is missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "package.json")
		err := os.WriteFile(path, []byte(`{"name": "test-pkg"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err = manifest.ReadPackageSpec(path)
		if !errors.Is(err, manifest.ErrMissingPackageManager) {
			t.Fatalf("expected ErrMissingPackageManager, got: %v", err)
		}
	})

	t.Run("return parsed spec when key exists", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "package.json")
		err := os.WriteFile(path, []byte(`{"packageManager": "pnpm@9.1.0+sha512.abcdef"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		spec, err := manifest.ReadPackageSpec(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if spec.Name != "pnpm" || spec.Version != "9.1.0" || spec.Hash != "+sha512.abcdef" {
			t.Fatalf("unexpected spec: %+v", spec)
		}
	})
}
