package ui_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

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
