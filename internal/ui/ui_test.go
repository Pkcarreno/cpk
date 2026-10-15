package ui_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pkcarreno/cpk/internal/manifest"
	"github.com/pkcarreno/cpk/internal/registry"
	"github.com/pkcarreno/cpk/internal/resolver"
	"github.com/pkcarreno/cpk/internal/ui"
)

func TestUI_StatusMessages(t *testing.T) {
	t.Run("routes status messages to stderr and keeps stdout empty", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		userInterface := ui.New(ui.Config{
			Out:     &stdout,
			Err:     &stderr,
			IsTTY:   true,
			NoColor: true,
		})

		userInterface.Success("installed %s", "pnpm@9.5.0")
		userInterface.Info("already up to date")
		userInterface.Warn("deprecated package")
		userInterface.Error("something failed")

		if stdout.Len() != 0 {
			t.Fatalf("expected stdout to be empty, got: %q", stdout.String())
		}

		errorOutput := stderr.String()
		expectedPhrases := []string{
			"✔ installed pnpm@9.5.0",
			"ℹ already up to date",
			"! deprecated package",
			"✖ something failed",
		}
		for _, phrase := range expectedPhrases {
			if !strings.Contains(errorOutput, phrase) {
				t.Fatalf("expected stderr to contain %q, got: %q", phrase, errorOutput)
			}
		}
	})

	t.Run("omits ansi color codes when color is disabled", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		userInterface := ui.New(ui.Config{
			Out:     &stdout,
			Err:     &stderr,
			IsTTY:   true,
			NoColor: true,
		})

		userInterface.Success("operation complete")

		if strings.Contains(stderr.String(), "\033[") {
			t.Fatalf("expected output without ANSI escape codes, got: %q", stderr.String())
		}
	})

	t.Run("writes ansi color codes when terminal and color are enabled", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		userInterface := ui.New(ui.Config{
			Out:     &stdout,
			Err:     &stderr,
			IsTTY:   true,
			NoColor: false,
		})

		userInterface.Success("operation complete")

		if !strings.Contains(stderr.String(), "\033[32m✔\033[0m") {
			t.Fatalf("expected green checkmark with ANSI codes, got: %q", stderr.String())
		}
	})

	t.Run("omits informational messages when quiet mode is enabled", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		userInterface := ui.New(ui.Config{
			Out:     &stdout,
			Err:     &stderr,
			IsTTY:   true,
			NoColor: true,
			IsQuiet: true,
		})

		userInterface.Success("saved successfully")
		userInterface.Info("fetching metadata")
		userInterface.Warn("low disk space")

		if stderr.Len() != 0 {
			t.Fatalf("expected stderr to be empty in quiet mode, got %q", stderr.String())
		}
	})

	t.Run("writes an error message when quiet mode is enabled", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		userInterface := ui.New(ui.Config{
			Out:     &stdout,
			Err:     &stderr,
			IsTTY:   true,
			NoColor: true,
		})

		userInterface.Error("fatal failure occurred")

		if !strings.Contains(stderr.String(), "fatal failure occurred") {
			t.Fatalf("expected stderr to contain error in quiet mode, got %q", stderr.String())
		}
	})
}

func TestUI_Spinner(t *testing.T) {
	t.Run("clears current line when stopping spinner on interactive terminal", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		userInterface := ui.New(ui.Config{
			Out:     &stdout,
			Err:     &stderr,
			IsTTY:   true,
			NoColor: false,
		})

		userInterface.StartSpinner("querying registry...")
		time.Sleep(50 * time.Millisecond)
		userInterface.StopSpinner()

		output := stderr.String()
		if !strings.Contains(output, "\r\033[K") {
			t.Fatalf("expected line clearing escape sequence \\r\\033[K, got: %q", output)
		}
	})

	t.Run("remains silent on a non-interactive terminal", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		userInterface := ui.New(ui.Config{
			Out:     &stdout,
			Err:     &stderr,
			IsTTY:   false,
			NoColor: true,
		})

		userInterface.StartSpinner("querying registry...")
		time.Sleep(50 * time.Millisecond)
		userInterface.StopSpinner()

		if stderr.Len() != 0 {
			t.Fatalf("expected non-interactive spinner to remain silent, got: %q", stderr.String())
		}
	})
}

func TestContext(t *testing.T) {
	t.Run("stores and retrieves the ui instance from context", func(t *testing.T) {
		var errorBuffer bytes.Buffer
		customUI := ui.New(ui.Config{Err: &errorBuffer})

		parentContext := ui.WithContext(context.Background(), customUI)
		retrievedUI := ui.FromContext(parentContext)

		if retrievedUI != customUI {
			t.Fatalf("expected retrieved UI %v, got %v", customUI, retrievedUI)
		}
	})

	t.Run("returns the default ui instance when context has no ui", func(t *testing.T) {
		retrievedUI := ui.FromContext(context.Background())
		if retrievedUI == nil {
			t.Fatal("expected non-nil default UI instance from empty context")
		}
	})
}

func TestEnvironmentDetection(t *testing.T) {
	t.Run("detects non-terminal for in-memory buffer", func(t *testing.T) {
		var outputBuffer bytes.Buffer
		if ui.IsTerminal(&outputBuffer) {
			t.Fatal("expected bytes.Buffer not to be reported as a terminal")
		}
	})

	t.Run("detects disabled colors when NO_COLOR environment variable is set", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		if !ui.HasNoColor() {
			t.Fatal("expected HasNoColor to report true when NO_COLOR is set")
		}
	})

	t.Run("detects continuous integration environment when CI variable is set", func(t *testing.T) {
		t.Setenv("CI", "true")
		if !ui.IsCI() {
			t.Fatal("expected IsCI to report true when CI is set")
		}
	})
}

func TestFormatError(t *testing.T) {
	t.Run("formats actionable error into error cause and fix fields", func(t *testing.T) {
		customErr := &ui.ActionableError{
			Summary: "sample summary",
			Cause:   "sample cause",
			Fix:     "sample fix",
		}

		result := ui.FormatError(customErr, false, true)

		if !strings.Contains(result, "Error: sample summary\n") {
			t.Fatalf("expected result to contain Error section, got: %q", result)
		}
		if !strings.Contains(result, "Cause: sample cause\n") {
			t.Fatalf("expected result to contain Cause section, got: %q", result)
		}
		if !strings.Contains(result, "Fix:   sample fix\n") {
			t.Fatalf("expected result to contain Fix section, got: %q", result)
		}
	})

	t.Run("formats actionable error with ansi colors when noColor is false", func(t *testing.T) {
		customErr := &ui.ActionableError{
			Summary: "sample summary",
			Cause:   "sample cause",
			Fix:     "sample fix",
		}

		result := ui.FormatError(customErr, false, false)

		if !strings.Contains(result, "\033[31mError:\033[0m sample summary\n") {
			t.Fatalf("expected result to contain colored Error section, got: %q", result)
		}
		if !strings.Contains(result, "\033[33mCause:\033[0m sample cause\n") {
			t.Fatalf("expected result to contain colored Cause section, got: %q", result)
		}
		if !strings.Contains(result, "\033[36mFix:\033[0m   sample fix\n") {
			t.Fatalf("expected result to contain colored Fix section, got: %q", result)
		}
	})

	t.Run("dumps full error details when verbose is true", func(t *testing.T) {
		underlyingErr := errors.New("dial tcp 127.0.0.1:443: connect: connection refused")
		customErr := &ui.ActionableError{
			Summary: "sample summary",
			Cause:   "sample cause",
			Fix:     "sample fix",
			Err:     underlyingErr,
		}

		result := ui.FormatError(customErr, true, true)

		if !strings.Contains(result, "Details:") {
			t.Fatalf("expected result to contain Details section, got: %q", result)
		}
		if !strings.Contains(result, "dial tcp 127.0.0.1:443: connect: connection refused") {
			t.Fatalf("expected result to contain underlying error details, got: %q", result)
		}
	})

	t.Run("omits error details when verbose is false", func(t *testing.T) {
		underlyingErr := errors.New("dial tcp 127.0.0.1:443: connect: connection refused")
		customErr := &ui.ActionableError{
			Summary: "sample summary",
			Cause:   "sample cause",
			Fix:     "sample fix",
			Err:     underlyingErr,
		}

		result := ui.FormatError(customErr, false, true)

		if strings.Contains(result, "Details:") {
			t.Fatalf("expected result not to contain Details section, got: %q", result)
		}
		if strings.Contains(result, "dial tcp 127.0.0.1:443: connect: connection refused") {
			t.Fatalf("expected result not to contain underlying error details, got: %q", result)
		}
	})
}

func TestFormatError_DomainErrors(t *testing.T) {
	testCases := []struct {
		name            string
		err             error
		expectedSummary string
		expectedCause   string
		expectedFix     string
	}{
		{
			name:            "manifest not found",
			err:             manifest.ErrManifestNotFound,
			expectedSummary: "package.json file not found",
			expectedCause:   "The package.json file does not exist in the current directory.",
			expectedFix:     "Run \"npm init -y\" to create a manifest, or go to a Node.js project directory.",
		},
		{
			name:            "missing package manager",
			err:             manifest.ErrMissingPackageManager,
			expectedSummary: "Missing packageManager field",
			expectedCause:   "The package.json file does not contain a packageManager field.",
			expectedFix:     "Run \"cpk use <package-manager>\" to set a package manager.",
		},
		{
			name:            "invalid package spec",
			err:             manifest.ErrInvalidPackageSpec,
			expectedSummary: "Invalid package specification",
			expectedCause:   "The packageManager field does not match the format name@version[+hash].",
			expectedFix:     "Run \"cpk use <package-manager>@<version>\" to set a valid package specification.",
		},
		{
			name:            "invalid manifest syntax",
			err:             manifest.ErrInvalidManifest,
			expectedSummary: "Invalid package.json manifest",
			expectedCause:   "The package.json file contains invalid JSON syntax.",
			expectedFix:     "Correct the JSON syntax errors in the package.json file.",
		},
		{
			name:            "package manager already exists",
			err:             manifest.ErrPackageManagerExists,
			expectedSummary: "packageManager field already exists",
			expectedCause:   "The package.json file already contains a packageManager field.",
			expectedFix:     "Run \"cpk use <package-manager>\" to update the existing packageManager field.",
		},
		{
			name:            "invalid integrity hash",
			err:             manifest.ErrInvalidIntegrityHash,
			expectedSummary: "Invalid integrity hash",
			expectedCause:   "The integrity hash format or digest length does not match supported algorithms.",
			expectedFix:     "Run \"cpk use <package-manager>\" to get a fresh hash from the registry.",
		},
		{
			name:            "unsupported package manager",
			err:             registry.ErrUnsupportedPackageManager,
			expectedSummary: "Unsupported package manager",
			expectedCause:   "cpk only supports pnpm, yarn, and npm.",
			expectedFix:     "Run \"cpk use <pnpm|yarn|npm>\" instead.",
		},
		{
			name:            "unsupported algorithm",
			err:             registry.ErrUnsupportedAlgorithm,
			expectedSummary: "Unsupported integrity algorithm",
			expectedCause:   "The registry integrity hash uses an unsupported algorithm.",
			expectedFix:     "Use a version with sha512, sha384, sha256, sha224, or sha1 hashes.",
		},
		{
			name:            "missing integrity in registry",
			err:             registry.ErrMissingIntegrity,
			expectedSummary: "Missing integrity hash in registry",
			expectedCause:   "The package metadata does not contain an integrity hash or shasum.",
			expectedFix:     "Select a package version that includes published integrity metadata.",
		},
		{
			name:            "package not found in registry",
			err:             registry.ErrPackageNotFound,
			expectedSummary: "Package not found in registry",
			expectedCause:   "The package does not exist in the configured npm registry.",
			expectedFix:     "Make sure the package name and registry URL are correct.",
		},
		{
			name:            "registry request failed",
			err:             registry.ErrRegistryRequestFailed,
			expectedSummary: "Registry request failed",
			expectedCause:   "The registry returned an error or the network request failed.",
			expectedFix:     "Check your network connection and registry URL.",
		},
		{
			name:            "version not found in registry",
			err:             resolver.ErrVersionNotFound,
			expectedSummary: "Version not found in registry",
			expectedCause:   "The registry does not have a release matching the target version.",
			expectedFix:     "Examine the package versions on npmjs.com or run \"cpk check\".",
		},
		{
			name:            "integrity mismatch",
			err:             errors.New("the integrity hash does not match the registry metadata"),
			expectedSummary: "Integrity hash mismatch",
			expectedCause:   "The integrity hash does not match the published registry metadata.",
			expectedFix:     "Run \"cpk use <package-manager>@<version>\" to update the manifest with the valid hash.",
		},
		{
			name:            "outdated package manager",
			err:             errors.New("the package manager is outdated"),
			expectedSummary: "Package manager is outdated",
			expectedCause:   "A newer version of the package manager exists in the registry.",
			expectedFix:     "Run \"cpk up\" to update to the latest version.",
		},
		{
			name:            "downgrade rejected",
			err:             errors.New("cannot downgrade with up: run 'cpk use' instead"),
			expectedSummary: "Downgrade rejected by up command",
			expectedCause:   "The target version is lower than the active pinned version.",
			expectedFix:     "Run \"cpk use <package-manager>@<version>\" to force a downgrade.",
		},
		{
			name:            "wrapped domain error",
			err:             fmt.Errorf("reading manifest: %w", manifest.ErrMissingPackageManager),
			expectedSummary: "Missing packageManager field",
			expectedCause:   "The package.json file does not contain a packageManager field.",
			expectedFix:     "Run \"cpk use <package-manager>\" to set a package manager.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := ui.FormatError(testCase.err, false, true)

			if !strings.Contains(result, testCase.expectedSummary) {
				t.Fatalf("expected result to contain colored Error section, got: %q", result)
			}
			if !strings.Contains(result, testCase.expectedCause) {
				t.Fatalf("expected result to contain colored Cause section, got: %q", result)
			}
			if !strings.Contains(result, testCase.expectedFix) {
				t.Fatalf("expected result to contain colored Fix section, got: %q", result)
			}
		})
	}
}
