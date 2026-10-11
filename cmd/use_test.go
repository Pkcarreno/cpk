package cmd_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkcarreno/cpk/cmd"
	"github.com/pkcarreno/cpk/internal/manifest"
	"github.com/pkcarreno/cpk/internal/registry"
)

func TestNewUseCommand(t *testing.T) {
	t.Run("rejects an unsupported package manager", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		err := os.WriteFile(manifestPath, []byte(`{"name": "test-pkg"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		var output bytes.Buffer

		useCommand := cmd.NewUseCommand()
		useCommand.SetArgs([]string{"bun"})
		useCommand.SetOut(&output)
		useCommand.SetErr(&output)

		err = useCommand.Execute()
		if !errors.Is(err, registry.ErrUnsupportedPackageManager) {
			t.Fatalf("expected ErrUnsupportedPackageManager, got: %v", err)
		}
	})

	testCases := []struct {
		name                  string
		arguments             []string
		expectedVersion       string
		expectedIntegrityHash string
	}{
		{
			name:                  "pins the package manager to the latest stable version when no target is specified",
			arguments:             []string{"pnpm"},
			expectedVersion:       "9.5.0",
			expectedIntegrityHash: "+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d",
		},
		{
			name:                  "pins the package manager to an explicit version when a target is specified",
			arguments:             []string{"pnpm@9.1.0"},
			expectedVersion:       "9.1.0",
			expectedIntegrityHash: "+sha512.01010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			temporalDirectory := t.TempDir()
			manifestPath := filepath.Join(temporalDirectory, "package.json")
			err := os.WriteFile(manifestPath, []byte(`{"name": "test-pkg"}`), 0600)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			t.Chdir(temporalDirectory)

			mockServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				responseWriter.Header().Set("Content-Type", "application/json")
				responseWriter.WriteHeader(http.StatusOK)
				responseWriter.Write([]byte(`{
						"name": "pnpm",
						"dist-tags": {"latest": "9.5.0"},
						"versions": {
							"9.1.0": {"dist": {"integrity": "sha512-AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ=="}},
							"9.5.0": {"dist": {"integrity": "sha512-hFbAm884LpT8E1I+sI2hcVojcCcUf+ylT5Yr34U3qRfH//2nAt7G1NFYmqGTDpRPnL1xZnMxIrA9/xuatbuZXQ=="}}
						}
					}`))
			}))
			defer mockServer.Close()

			t.Setenv("NPM_CONFIG_REGISTRY", mockServer.URL)

			var output bytes.Buffer

			useCommand := cmd.NewUseCommand()
			useCommand.SetArgs(testCase.arguments)
			useCommand.SetOut(&output)
			useCommand.SetErr(&output)

			err = useCommand.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			expectedOutput := ""

			if output.String() != expectedOutput {
				t.Fatalf("output does not match expected output, expected: %q, got: %v", expectedOutput, output)
			}

			updatedPackageSpec, err := manifest.ReadPackageSpec(manifestPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			expectedPackageSpec := manifest.PackageSpec{
				Name:    "pnpm",
				Version: testCase.expectedVersion,
				Hash:    testCase.expectedIntegrityHash,
			}

			if updatedPackageSpec != expectedPackageSpec {
				t.Fatalf("updated package spec does not match expected package spec, expected: %q, got: %q", expectedPackageSpec.String(), updatedPackageSpec.String())
			}
		})
	}

	t.Run("preserves the manifest file when the version and integrity hash already match", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		initialContent := []byte(`{"name": "test-pkg","packageManager": "pnpm@9.5.0+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d"}`)
		err := os.WriteFile(manifestPath, initialContent, 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		mockServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.Header().Set("Content-Type", "application/json")
			responseWriter.WriteHeader(http.StatusOK)
			responseWriter.Write([]byte(`{
					"name": "pnpm",
					"dist-tags": {"latest": "9.5.0"},
					"versions": {
						"9.1.0": {"dist": {"integrity": "sha512-AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ=="}},
						"9.5.0": {"dist": {"integrity": "sha512-hFbAm884LpT8E1I+sI2hcVojcCcUf+ylT5Yr34U3qRfH//2nAt7G1NFYmqGTDpRPnL1xZnMxIrA9/xuatbuZXQ=="}}
					}
				}`))
		}))
		defer mockServer.Close()

		t.Setenv("NPM_CONFIG_REGISTRY", mockServer.URL)

		var output bytes.Buffer

		useCommand := cmd.NewUseCommand()
		useCommand.SetArgs([]string{"pnpm@9.5.0"})
		useCommand.SetOut(&output)
		useCommand.SetErr(&output)

		beforeFileInfo, err := os.Stat(manifestPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = useCommand.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		afterFileInfo, err := os.Stat(manifestPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !afterFileInfo.ModTime().Equal(beforeFileInfo.ModTime()) {
			t.Fatalf("expected the file modification time to remain unchanged")
		}

		expectedOutput := ""

		if output.String() != expectedOutput {
			t.Fatalf("output does not match expected output, expected: %q, got: %v", expectedOutput, output)
		}

		updatedPackageSpec, err := manifest.ReadPackageSpec(manifestPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedPackageSpec := manifest.PackageSpec{
			Name:    "pnpm",
			Version: "9.5.0",
			Hash:    "+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d",
		}

		if updatedPackageSpec != expectedPackageSpec {
			t.Fatalf("updated package spec does not match expected package spec, expected: %q, got: %q", expectedPackageSpec.String(), updatedPackageSpec.String())
		}
	})

	t.Run("returns an error when the package.json file does not exist", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		t.Chdir(temporalDirectory)

		var output bytes.Buffer

		useCommand := cmd.NewUseCommand()
		useCommand.SetArgs([]string{"pnpm"})
		useCommand.SetOut(&output)
		useCommand.SetErr(&output)

		err := useCommand.Execute()
		if !errors.Is(err, manifest.ErrManifestNotFound) {
			t.Fatalf("expected ErrManifestNotFound, got: %v", err)
		}
	})
}
