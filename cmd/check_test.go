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

func TestNewCheckCommand(t *testing.T) {
	t.Run("fails when the package.json file does not exist", func(t *testing.T) {
		temporalDirectory := t.TempDir()

		t.Chdir(temporalDirectory)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		checkCommand := cmd.NewCheckCommand()
		checkCommand.SetOut(&stdout)
		checkCommand.SetErr(&stderr)

		err := checkCommand.Execute()
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

		checkCommand := cmd.NewCheckCommand()
		checkCommand.SetOut(&stdout)
		checkCommand.SetErr(&stderr)

		err = checkCommand.Execute()
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

		checkCommand := cmd.NewCheckCommand()
		checkCommand.SetOut(&stdout)
		checkCommand.SetErr(&stderr)

		err = checkCommand.Execute()
		if !errors.Is(err, registry.ErrUnsupportedPackageManager) {
			t.Fatalf("expected ErrUnsupportedPackageManager, got: %v", err)
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

		checkCommand := cmd.NewCheckCommand()
		checkCommand.SetOut(&stdout)
		checkCommand.SetErr(&stderr)
		checkCommand.SetContext(ui.WithContext(context.Background(), uiInstance))

		err = checkCommand.Execute()
		if !errors.Is(err, cmd.ErrIntegrityMismatch) {
			t.Fatalf("expected ErrIntegrityMismatch, got: %v", err)
		}
	})

	t.Run("reports active package manager is up to date", func(t *testing.T) {
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

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		uiInstance := ui.New(ui.Config{
			Out:     &stdout,
			Err:     &stderr,
			NoColor: true,
		})

		checkCommand := cmd.NewCheckCommand()
		checkCommand.SetOut(&stdout)
		checkCommand.SetErr(&stderr)
		checkCommand.SetContext(ui.WithContext(context.Background(), uiInstance))

		err = checkCommand.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stdout.String() != "" {
			t.Fatalf("expected stdout to be empty, got: %q", stdout.String())
		}

		expectedStderr := fmt.Sprintf("ℹ %s is already up to date (%s)\n", "pnpm", "9.5.0")

		if stderr.String() != expectedStderr {
			t.Fatalf("stderr does not match expected output, expected: %q, got: %q", expectedStderr, stderr.String())
		}
	})

	t.Run("reports active package manager has an available update", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		initialContent := []byte(`{"name": "test-pkg","packageManager": "pnpm@9.1.0+sha512.01010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101"}`)
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

		checkCommand := cmd.NewCheckCommand()
		checkCommand.SetOut(&stdout)
		checkCommand.SetErr(&stderr)
		checkCommand.SetContext(ui.WithContext(context.Background(), uiInstance))

		err = checkCommand.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stdout.String() != "" {
			t.Fatalf("expected stdout to be empty, got: %q", stdout.String())
		}

		expectedStderr := fmt.Sprintf("! Update available for pnpm: 9.1.0 -> 9.5.0\n")

		if stderr.String() != expectedStderr {
			t.Fatalf("stderr does not match expected output, expected: %q, got: %q", expectedStderr, stderr.String())
		}
	})

	t.Run("fails when outdated and fail-on-outdated is set", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		initialContent := []byte(`{"name": "test-pkg","packageManager": "pnpm@9.1.0+sha512.01010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101"}`)
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

		checkCommand := cmd.NewCheckCommand()
		checkCommand.SetArgs([]string{"--fail-on-outdated"})
		checkCommand.SetOut(&stdout)
		checkCommand.SetErr(&stderr)
		checkCommand.SetContext(ui.WithContext(context.Background(), uiInstance))

		err = checkCommand.Execute()
		if !errors.Is(err, cmd.ErrOutdatedPackage) {
			t.Fatalf("expected ErrOutdatedPackage, got: %v", err)
		}
	})
}
