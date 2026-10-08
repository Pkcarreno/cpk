package resolver_test

import (
	"errors"
	"testing"

	"github.com/pkcarreno/cpk/internal/registry"
	"github.com/pkcarreno/cpk/internal/resolver"
)

func TestResolveVersion(t *testing.T) {
	t.Run("resolves dist-tag to a canonical version", func(t *testing.T) {
		metadata := registry.PackageMetadata{
			Name: "pnpm",
			DistTags: map[string]string{
				"latest": "9.1.0",
			},
			Versions: map[string]registry.PackageVersion{
				"9.1.0": {},
			},
		}

		version, err := resolver.ResolveVersion(metadata, "latest")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if version != "9.1.0" {
			t.Fatalf("expected version %q, got: %q", "9.1.0", version)
		}
	})

	t.Run("returns ErrVersionNotFound when query is empty", func(t *testing.T) {
		metadata := registry.PackageMetadata{
			DistTags: map[string]string{
				"latest": "9.1.0",
			},
			Versions: map[string]registry.PackageVersion{
				"9.1.0": {},
			},
		}

		_, err := resolver.ResolveVersion(metadata, "")
		if !errors.Is(err, resolver.ErrVersionNotFound) {
			t.Fatalf("expected ErrVersionNotFound, got: %v", err)
		}
	})

	t.Run("resolves an exact version from package metadata", func(t *testing.T) {
		metadata := registry.PackageMetadata{
			Versions: map[string]registry.PackageVersion{
				"9.1.0": {},
				"8.5.0": {},
			},
		}

		version, err := resolver.ResolveVersion(metadata, "8.5.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if version != "8.5.0" {
			t.Fatalf("expected version %q, got: %q", "8.5.0", version)
		}
	})

	t.Run("returns ErrVersionNotFound when the version does not exist", func(t *testing.T) {
		metadata := registry.PackageMetadata{
			DistTags: map[string]string{
				"latest": "9.1.0",
			},
			Versions: map[string]registry.PackageVersion{
				"9.1.0": {},
			},
		}

		_, err := resolver.ResolveVersion(metadata, "10.0.0")
		if !errors.Is(err, resolver.ErrVersionNotFound) {
			t.Fatalf("expected ErrVersionNotFound, got: %v", err)
		}
	})

	t.Run("resolves a single-digit numeric prefix to the highest version", func(t *testing.T) {
		metadata := registry.PackageMetadata{
			Versions: map[string]registry.PackageVersion{
				"8.9.0":  {},
				"9.1.0":  {},
				"9.2.1":  {},
				"9.10.0": {},
				"10.0.0": {},
			},
		}

		version, err := resolver.ResolveVersion(metadata, "9")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if version != "9.10.0" {
			t.Fatalf("expected version %q, got: %q", "9.10.0", version)
		}
	})

	t.Run("resolves a major-minor numeric prefix to the highest version", func(t *testing.T) {
		metadata := registry.PackageMetadata{
			Versions: map[string]registry.PackageVersion{
				"9.2.0":  {},
				"9.2.5":  {},
				"9.2.12": {},
				"9.3.0":  {},
			},
		}

		version, err := resolver.ResolveVersion(metadata, "9.2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if version != "9.2.12" {
			t.Fatalf("expected version %q, got: %q", "9.2.12", version)
		}
	})

	t.Run("ignores prerelease versions when matching a numeric prefix", func(t *testing.T) {
		metadata := registry.PackageMetadata{
			Versions: map[string]registry.PackageVersion{
				"9.2.0":      {},
				"9.3.0-rc.1": {},
			},
		}

		version, err := resolver.ResolveVersion(metadata, "9")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if version != "9.2.0" {
			t.Fatalf("expected version %q, got: %q", "9.2.0", version)
		}
	})
}
