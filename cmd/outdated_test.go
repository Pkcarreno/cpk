package cmd_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkcarreno/cpk/cmd"
	"github.com/pkcarreno/cpk/internal/manifest"
	"github.com/pkcarreno/cpk/internal/registry"
	"github.com/pkcarreno/cpk/internal/ui"
)

func TestNewOutdatedCommand(t *testing.T) {
	t.Run("fails when the package.json file does not exist", func(t *testing.T) {
		temporalDirectory := t.TempDir()

		t.Chdir(temporalDirectory)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		outdatedCommand := cmd.NewOutdatedCommand()
		outdatedCommand.SetOut(&stdout)
		outdatedCommand.SetErr(&stderr)

		err := outdatedCommand.Execute()
		if !errors.Is(err, manifest.ErrManifestNotFound) {
			t.Fatalf("expected ErrManifestNotFound, got: %v", err)
		}
	})

	t.Run("fails when the package.json file has no packageManager field", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		err := os.WriteFile(manifestPath, []byte(`{"name": "test-pkg"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		outdatedCommand := cmd.NewOutdatedCommand()
		outdatedCommand.SetOut(&stdout)
		outdatedCommand.SetErr(&stderr)

		err = outdatedCommand.Execute()
		if !errors.Is(err, manifest.ErrMissingPackageManager) {
			t.Fatalf("expected ErrMissingPackageManager, got: %v", err)
		}
	})

	t.Run("rejects an unsupported package manager", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		err := os.WriteFile(manifestPath, []byte(`{"name": "test-pkg", "packageManager": "bun@1.4.5"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		outdatedCommand := cmd.NewOutdatedCommand()
		outdatedCommand.SetOut(&stdout)
		outdatedCommand.SetErr(&stderr)

		err = outdatedCommand.Execute()
		if !errors.Is(err, registry.ErrUnsupportedPackageManager) {
			t.Fatalf("expected ErrUnsupportedPackageManager, got: %v", err)
		}
	})

	testCases := []struct {
		name           string
		specString     string
		arguments      []string
		expectedStderr string
		expectedErr    error
	}{
		{
			name:           "reports active package manager is up to date without hash",
			specString:     "pnpm@9.5.0",
			arguments:      []string{},
			expectedStderr: "ℹ pnpm is already up to date (9.5.0)\n",
			expectedErr:    nil,
		},
		{
			name:           "reports active package manager is up to date with valid sha512 hash",
			specString:     "pnpm@9.5.0+sha512." + strings.Repeat("c", 128),
			arguments:      []string{},
			expectedStderr: "ℹ pnpm is already up to date (9.5.0)\n",
			expectedErr:    nil,
		},
		{
			name:           "reports active package manager is up to date with corrupted hash",
			specString:     "pnpm@9.5.0+corrupted_hash",
			arguments:      []string{},
			expectedStderr: "ℹ pnpm is already up to date (9.5.0)\n",
			expectedErr:    nil,
		},
		{
			name:           "reports Yarn Classic is up to date on 1.x line",
			specString:     "yarn@1.22.22+sha1." + strings.Repeat("a", 40),
			arguments:      []string{},
			expectedStderr: "ℹ yarn is already up to date (1.22.22)\n",
			expectedErr:    nil,
		},
		{
			name:           "reports Yarn Berry is up to date",
			specString:     "yarn@4.18.0",
			arguments:      []string{},
			expectedStderr: "ℹ yarn is already up to date (4.18.0)\n",
			expectedErr:    nil,
		},
		{
			name:           "reports update available when version is lower with valid hash",
			specString:     "pnpm@9.1.0+sha256." + strings.Repeat("b", 64),
			arguments:      []string{},
			expectedStderr: "! Update available for pnpm: 9.1.0 -> 9.5.0\n",
			expectedErr:    nil,
		},
		{
			name:           "reports update available when version is lower with corrupted hash",
			specString:     "pnpm@9.1.0+sha512.invalid",
			arguments:      []string{},
			expectedStderr: "! Update available for pnpm: 9.1.0 -> 9.5.0\n",
			expectedErr:    nil,
		},
		{
			name:           "reports Yarn Classic update available within 1.x",
			specString:     "yarn@1.22.19",
			arguments:      []string{},
			expectedStderr: "! Update available for yarn: 1.22.19 -> 1.22.22\n",
			expectedErr:    nil,
		},
		{
			name:           "reports Yarn Berry update available",
			specString:     "yarn@4.1.0",
			arguments:      []string{},
			expectedStderr: "! Update available for yarn: 4.1.0 -> 4.18.0\n",
			expectedErr:    nil,
		},
		{
			name:           "fails when update is available and fail-on-outdated flag is set",
			specString:     "pnpm@9.1.0",
			arguments:      []string{"--fail-on-outdated"},
			expectedStderr: "! Update available for pnpm: 9.1.0 -> 9.5.0\n",
			expectedErr:    cmd.ErrOutdatedPackage,
		},
		{
			name:           "succeeds when up to date and fail-on-outdated flag is set",
			specString:     "pnpm@9.5.0",
			arguments:      []string{"--fail-on-outdated"},
			expectedStderr: "ℹ pnpm is already up to date (9.5.0)\n",
			expectedErr:    nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			temporalDirectory := t.TempDir()
			manifestPath := filepath.Join(temporalDirectory, "package.json")
			initialContent := []byte(`{"name": "test-pkg", "packageManager": "` + testCase.specString + `"}`)
			err := os.WriteFile(manifestPath, initialContent, 0600)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			t.Chdir(temporalDirectory)

			mockServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				responseWriter.Header().Set("Content-Type", "application/json")
				switch request.URL.Path {
				case "/pnpm":
					responseWriter.WriteHeader(http.StatusOK)
					responseWriter.Write([]byte(`{
		                "name": "pnpm",
		                "dist-tags": {"latest": "9.5.0"},
		                "versions": {
		                    "9.1.0": {"dist": {"integrity": "sha512-AQ=="}},
		                    "9.5.0": {"dist": {"integrity": "sha512-hF=="}}
		                }
		            }`))
				case "/yarn":
					responseWriter.WriteHeader(http.StatusOK)
					responseWriter.Write([]byte(`{
		                "name": "yarn",
		                "dist-tags": {"latest": "1.22.22"},
		                "versions": {
		                    "1.22.19": {"dist": {"shasum": "010101"}},
		                    "1.22.22": {"dist": {"shasum": "ac3454"}},
		                    "2.0.0": {"dist": {"shasum": "020202"}}
		                }
		            }`))
				case "/@yarnpkg/cli-dist":
					responseWriter.WriteHeader(http.StatusOK)
					responseWriter.Write([]byte(`{
		                "name": "@yarnpkg/cli-dist",
		                "dist-tags": {"latest": "4.18.0"},
		                "versions": {
		                    "4.1.0": {"dist": {"integrity": "sha512-AQ=="}},
		                    "4.18.0": {"dist": {"integrity": "sha512-hF=="}}
		                }
		            }`))
				default:
					responseWriter.WriteHeader(http.StatusNotFound)
				}
			}))
			defer mockServer.Close()

			t.Setenv("NPM_CONFIG_REGISTRY", mockServer.URL)

			var stdout bytes.Buffer
			var stderr bytes.Buffer

			uiInstance := ui.New(ui.Config{
				Out:     &stdout,
				Err:     &stderr,
				NoColor: true,
			})

			outdatedCommand := cmd.NewOutdatedCommand()
			outdatedCommand.SetArgs(testCase.arguments)
			outdatedCommand.SetOut(&stdout)
			outdatedCommand.SetErr(&stderr)
			outdatedCommand.SetContext(ui.WithContext(context.Background(), uiInstance))

			err = outdatedCommand.Execute()
			if testCase.expectedErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if testCase.expectedErr != nil && !errors.Is(err, testCase.expectedErr) {
				t.Fatalf("expected error: %v, got: %v", testCase.expectedErr, err)
			}

			if stdout.String() != "" {
				t.Fatalf("expected stdout to be empty, got: %q", stdout.String())
			}

			if stderr.String() != testCase.expectedStderr {
				t.Fatalf("stderr does not match expected output, expected: %q, got: %q", testCase.expectedStderr, stderr.String())
			}

			finalContent, _ := os.ReadFile(manifestPath)
			if !bytes.Equal(finalContent, initialContent) {
				t.Fatal("expected manifest to remain unmodified on disk")
			}
		})
	}
}
