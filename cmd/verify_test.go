package cmd_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkcarreno/cpk/cmd"
	"github.com/pkcarreno/cpk/internal/manifest"
	"github.com/pkcarreno/cpk/internal/registry"
	"github.com/pkcarreno/cpk/internal/ui"
)

func TestNewVerifyCommand(t *testing.T) {
	t.Run("fails when the package.json file does not exist", func(t *testing.T) {
		temporalDirectory := t.TempDir()

		t.Chdir(temporalDirectory)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		verifyCommand := cmd.NewVerifyCommand()
		verifyCommand.SetOut(&stdout)
		verifyCommand.SetErr(&stderr)

		err := verifyCommand.Execute()
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

		verifyCommand := cmd.NewVerifyCommand()
		verifyCommand.SetOut(&stdout)
		verifyCommand.SetErr(&stderr)

		err = verifyCommand.Execute()
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

		verifyCommand := cmd.NewVerifyCommand()
		verifyCommand.SetOut(&stdout)
		verifyCommand.SetErr(&stderr)

		err = verifyCommand.Execute()
		if !errors.Is(err, registry.ErrUnsupportedPackageManager) {
			t.Fatalf("expected ErrUnsupportedPackageManager, got: %v", err)
		}
	})

	t.Run("fails when manifest does not contain an integrity hash", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		err := os.WriteFile(manifestPath, []byte(`{"name": "test-pkg", "packageManager": "pnpm@9.5.0"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		verifyCommand := cmd.NewVerifyCommand()
		verifyCommand.SetOut(&stdout)
		verifyCommand.SetErr(&stderr)

		err = verifyCommand.Execute()
		if !errors.Is(err, cmd.ErrMissingIntegrityHash) {
			t.Fatalf("expected ErrMissingIntegrityHash, got: %v", err)
		}

		var actionableError *ui.ActionableError
		if !errors.As(err, &actionableError) {
			t.Fatalf("expected ActionableError, got: %T", err)
		}
		expectedErrorFix := fmt.Sprintf(`Run "cpk use %s@%s"`, "pnpm", "9.5.0")
		if actionableError.Fix != expectedErrorFix {
			t.Fatalf("expected fix %q, got: %q", expectedErrorFix, actionableError.Fix)
		}
	})

	t.Run("fails when manifest integrity does not match registry", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		initialContent := []byte(`{"name": "test-pkg","packageManager": "pnpm@9.1.0+sha512.11010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101"}`)
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

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		uiInstance := ui.New(ui.Config{
			Out:     &stdout,
			Err:     &stderr,
			NoColor: true,
		})

		verifyCommand := cmd.NewVerifyCommand()
		verifyCommand.SetOut(&stdout)
		verifyCommand.SetErr(&stderr)
		verifyCommand.SetContext(ui.WithContext(context.Background(), uiInstance))

		err = verifyCommand.Execute()
		if !errors.Is(err, cmd.ErrIntegrityMismatch) {
			t.Fatalf("expected ErrIntegrityMismatch, got: %v", err)
		}

		var actionableError *ui.ActionableError
		if !errors.As(err, &actionableError) {
			t.Fatalf("expected ActionableError, got: %T", err)
		}
		expectedErrorFix := fmt.Sprintf(`Run "cpk use %s@%s"`, "pnpm", "9.1.0")
		if actionableError.Fix != expectedErrorFix {
			t.Fatalf("expected fix %q, got: %q", expectedErrorFix, actionableError.Fix)
		}

		finalContent, _ := os.ReadFile(manifestPath)
		if !bytes.Equal(finalContent, initialContent) {
			t.Fatal("expected manifest to remain unmodified on disk")
		}
	})

	testCases := []struct {
		name           string
		specString     string
		endpoint       string
		version        string
		distJSON       string
		expectedStderr string
	}{
		{
			name:           "verifies sha512 integrity hash",
			specString:     "pnpm@9.5.0+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d",
			endpoint:       "/pnpm",
			version:        "9.5.0",
			distJSON:       `{"integrity": "sha512-hFbAm884LpT8E1I+sI2hcVojcCcUf+ylT5Yr34U3qRfH//2nAt7G1NFYmqGTDpRPnL1xZnMxIrA9/xuatbuZXQ=="}`,
			expectedStderr: "✔ Integrity hash verified for pnpm@9.5.0\n",
		},
		{
			name:           "verifies sha384 integrity hash",
			specString:     "pnpm@9.5.0+sha384.fdbd8e75a67f29f701a4e040385e2e23986303ea10239211af907fcbb83578b3e417cb71ce646efd0819dd8c088de1bd",
			endpoint:       "/pnpm",
			version:        "9.5.0",
			distJSON:       `{"integrity": "sha384-/b2OdaZ/KfcBpOBAOF4uI5hjA+oQI5IRr5B/y7g1eLPkF8txzmRu/QgZ3YwIjeG9"}`,
			expectedStderr: "✔ Integrity hash verified for pnpm@9.5.0\n",
		},
		{
			name:           "verifies sha256 integrity hash",
			specString:     "pnpm@9.5.0+sha256.b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			endpoint:       "/pnpm",
			version:        "9.5.0",
			distJSON:       `{"integrity": "sha256-uU0nuZNNPgilLlLX2n2r+sSE7+N6U4DukIj3rOLvzek="}`,
			expectedStderr: "✔ Integrity hash verified for pnpm@9.5.0\n",
		},
		{
			name:           "verifies sha1 integrity hash for yarn classic",
			specString:     "yarn@1.22.22+sha1.2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
			endpoint:       "/yarn",
			version:        "1.22.22",
			distJSON:       `{"integrity": "sha1-Kq5sNclPz7QV2+lfQIuc6R7oRu0="}`,
			expectedStderr: "✔ Integrity hash verified for yarn@1.22.22\n",
		},
		{
			name:           "verifies legacy shasum fallback as sha1",
			specString:     "yarn@1.22.19+sha1.2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
			endpoint:       "/yarn",
			version:        "1.22.19",
			distJSON:       `{"shasum": "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"}`,
			expectedStderr: "✔ Integrity hash verified for yarn@1.22.19\n",
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
				if request.URL.Path != testCase.endpoint {
					responseWriter.WriteHeader(http.StatusNotFound)
					return
				}
				responseWriter.Header().Set("Content-Type", "application/json")
				responseWriter.WriteHeader(http.StatusOK)
				payload := fmt.Sprintf(`{"versions": {%q: {"dist": %s}}}`, testCase.version, testCase.distJSON)
				responseWriter.Write([]byte(payload))
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

			verifyCommand := cmd.NewVerifyCommand()
			verifyCommand.SetOut(&stdout)
			verifyCommand.SetErr(&stderr)
			verifyCommand.SetContext(ui.WithContext(context.Background(), uiInstance))

			err = verifyCommand.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
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
