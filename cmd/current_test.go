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

func TestNewCurrentCommand(t *testing.T) {
	t.Run("prints the active package manager specification", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		err := os.WriteFile(manifestPath, []byte(`{"name": "test-pkg", "packageManager": "pnpm@9.1.0+sha512.abcdef"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		var output bytes.Buffer

		currentCommand := cmd.NewCurrentCommand()
		currentCommand.SetOut(&output)
		currentCommand.SetErr(&output)

		err = currentCommand.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedOutput := "Package Manager:\tpnpm\nVersion:        \t9.1.0\nIntegrity Hash: \t+sha512.abcdef\n"

		if output.String() != expectedOutput {
			t.Fatalf("output does not match expected output, expected: %q, got: %v", expectedOutput, output)
		}
	})

	t.Run("returns an error when package.json does not exist", func(t *testing.T) {
		temporalDirectory := t.TempDir()

		t.Chdir(temporalDirectory)

		var output bytes.Buffer

		currentCommand := cmd.NewCurrentCommand()
		currentCommand.SetOut(&output)
		currentCommand.SetErr(&output)

		err := currentCommand.Execute()
		if !errors.Is(err, manifest.ErrManifestNotFound) {
			t.Fatalf("expected ErrManifestNotFound, got: %v", err)
		}
	})

	t.Run("returns an error when packageManager field is missing", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		err := os.WriteFile(manifestPath, []byte(`{"name": "test-pkg"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		var output bytes.Buffer

		currentCommand := cmd.NewCurrentCommand()
		currentCommand.SetOut(&output)
		currentCommand.SetErr(&output)

		err = currentCommand.Execute()
		if !errors.Is(err, manifest.ErrMissingPackageManager) {
			t.Fatalf("expected ErrMissingPackageManager, got: %v", err)
		}
	})

	t.Run("prints the exact specification string when raw flag is passed", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		err := os.WriteFile(manifestPath, []byte(`{"name": "test-pkg", "packageManager": "pnpm@9.1.0+sha512.abcdef"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		var output bytes.Buffer

		currentCommand := cmd.NewCurrentCommand()
		currentCommand.SetArgs([]string{"--raw"})
		currentCommand.SetOut(&output)
		currentCommand.SetErr(&output)

		err = currentCommand.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedOutput := "pnpm@9.1.0+sha512.abcdef\n"

		if output.String() != expectedOutput {
			t.Fatalf("output does not match expected output, expected: %q, got: %v", expectedOutput, output)
		}
	})

	t.Run("indicates an available update when check flag is passed", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		err := os.WriteFile(manifestPath, []byte(`{"name": "test-pkg", "packageManager": "pnpm@9.1.0+sha512.abcdef"}`), 0600)
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
						"9.1.0": {"dist": {"integrity": "sha512-abcdef"}},
						"9.5.0": {"dist": {"integrity": "sha512-latest"}}
					}
				}`))
		}))
		defer mockServer.Close()

		t.Setenv("NPM_CONFIG_REGISTRY", mockServer.URL)

		var output bytes.Buffer

		currentCommand := cmd.NewCurrentCommand()
		currentCommand.SetArgs([]string{"--check"})
		currentCommand.SetOut(&output)
		currentCommand.SetErr(&output)

		err = currentCommand.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedOutput := "Package Manager:\tpnpm\nVersion:        \t9.1.0\nIntegrity Hash: \t+sha512.abcdef\n\nUpdate available: 9.1.0 -> 9.5.0\n"

		if output.String() != expectedOutput {
			t.Fatalf("output does not match expected output, expected: %q, got: %v", expectedOutput, output)
		}
	})

	t.Run("indicates no updates available when the version is current", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		err := os.WriteFile(manifestPath, []byte(`{"name": "test-pkg", "packageManager": "pnpm@9.1.0+sha512.abcdef"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		mockServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.Header().Set("Content-Type", "application/json")
			responseWriter.WriteHeader(http.StatusOK)
			responseWriter.Write([]byte(`{
					"name": "pnpm",
					"dist-tags": {"latest": "9.1.0"},
					"versions": {
						"9.1.0": {"dist": {"integrity": "sha512-abcdef"}}
					}
				}`))
		}))
		defer mockServer.Close()

		t.Setenv("NPM_CONFIG_REGISTRY", mockServer.URL)

		var output bytes.Buffer

		currentCommand := cmd.NewCurrentCommand()
		currentCommand.SetArgs([]string{"--check"})
		currentCommand.SetOut(&output)
		currentCommand.SetErr(&output)

		err = currentCommand.Execute()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedOutput := "Package Manager:\tpnpm\nVersion:        \t9.1.0\nIntegrity Hash: \t+sha512.abcdef\n\nNo updates available\n"

		if output.String() != expectedOutput {
			t.Fatalf("output does not match expected output, expected: %q, got: %v", expectedOutput, output)
		}
	})

	t.Run("returns an error when the package manager is not supported during check", func(t *testing.T) {
		temporalDirectory := t.TempDir()
		manifestPath := filepath.Join(temporalDirectory, "package.json")
		err := os.WriteFile(manifestPath, []byte(`{"name": "test-pkg", "packageManager": "bun@1.0.0"}`), 0600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		t.Chdir(temporalDirectory)

		var output bytes.Buffer
		currentCommand := cmd.NewCurrentCommand()
		currentCommand.SetArgs([]string{"--check"})
		currentCommand.SetOut(&output)
		currentCommand.SetErr(&output)

		err = currentCommand.Execute()
		if !errors.Is(err, registry.ErrUnsupportedPackageManager) {
			t.Fatalf("expected ErrUnsupportedPackageManager, got: %v", err)
		}
	})
}
