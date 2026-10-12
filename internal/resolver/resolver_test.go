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

func TestIsLowerVersion(t *testing.T) {
	testCases := []struct {
		name             string
		candidateVersion string
		currentVersion   string
		isLowerExpected  bool
	}{
		{
			name:             "returns true when the candidate version is lower than the current version",
			candidateVersion: "9.2.4",
			currentVersion:   "9.5.0",
			isLowerExpected:  true,
		},
		{
			name:             "returns false when the candidate version is higher than the current version",
			candidateVersion: "9.5.0",
			currentVersion:   "9.2.4",
			isLowerExpected:  false,
		},
		{
			name:             "returns false when the candidate version is equal to the current version",
			candidateVersion: "9.5.0",
			currentVersion:   "9.5.0",
			isLowerExpected:  false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			isLower, err := resolver.IsLowerVersion(testCase.candidateVersion, testCase.currentVersion)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if isLower != testCase.isLowerExpected {
				t.Fatalf("expected value does not match the returned value, expected: %v, got: %v", testCase.isLowerExpected, isLower)
			}

		})
	}
}

func TestResolvePackageName(t *testing.T) {
	testCases := []struct {
		name           string
		packageManager string
		targetVersion  string
		expectedName   string
	}{
		{
			name:           "returns the package manager name when package manager is not yarn",
			packageManager: "pnpm",
			targetVersion:  "latest",
			expectedName:   "pnpm",
		},
		{
			name:           "returns berry package name when yarn target is latest",
			packageManager: "yarn",
			targetVersion:  "latest",
			expectedName:   "@yarnpkg/cli-dist",
		},
		{
			name:           "returns berry package name when yarn target is empty",
			packageManager: "yarn",
			targetVersion:  "",
			expectedName:   "@yarnpkg/cli-dist",
		},
		{
			name:           "returns classic yarn package name when target version is lower than 2.0.0",
			packageManager: "yarn",
			targetVersion:  "1.22.22",
			expectedName:   "yarn",
		},
		{
			name:           "returns classic yarn package name when target version prefix is 1",
			packageManager: "yarn",
			targetVersion:  "1",
			expectedName:   "yarn",
		},
		{
			name:           "returns classic yarn package name when target version prefix is 1.22",
			packageManager: "yarn",
			targetVersion:  "1.22",
			expectedName:   "yarn",
		},
		{
			name:           "returns berry package name when target version is 2.0.0",
			packageManager: "yarn",
			targetVersion:  "2.0.0",
			expectedName:   "@yarnpkg/cli-dist",
		},
		{
			name:           "returns berry package name when target version is higher than 2.0.0",
			packageManager: "yarn",
			targetVersion:  "4.18.0",
			expectedName:   "@yarnpkg/cli-dist",
		},
		{
			name:           "returns berry package name when target version prefix is 4",
			packageManager: "yarn",
			targetVersion:  "4",
			expectedName:   "@yarnpkg/cli-dist",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualName := resolver.ResolvePackageName(testCase.packageManager, testCase.targetVersion)
			if actualName != testCase.expectedName {
				t.Fatalf("expected package name %q, got: %q", testCase.expectedName, actualName)
			}
		})
	}
}
