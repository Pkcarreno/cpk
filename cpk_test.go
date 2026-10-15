package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun_ErrorFormattingAndExitCode(t *testing.T) {
	t.Run("formats actionable error on stderr and returns exit code 1", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		exitCode := run([]string{"use", "unsupported-tool", "--no-color"}, &stdout, &stderr)

		if exitCode != 1 {
			t.Fatalf("expected exit code 1, got %d", exitCode)
		}
		if stdout.Len() != 0 {
			t.Fatalf("expected stdout to be empty, got: %q", stdout.String())
		}

		output := stderr.String()
		expectedSections := []string{
			"Error: package.json file not found",
			"Cause: The package.json file does not exist in the current directory.",
			"Fix:   Run \"npm init -y\" to create a manifest, or go to a Node.js project directory.",
		}
		for _, section := range expectedSections {
			if !strings.Contains(output, section) {
				t.Fatalf("expected stderr to contain %q, got: %q", section, output)
			}
		}
	})

	t.Run("dumps error details when verbose flag is passed", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		exitCode := run([]string{"use", "unsupported-tool", "--verbose", "--no-color"}, &stdout, &stderr)

		if exitCode != 1 {
			t.Fatalf("expected exit code 1, got %d", exitCode)
		}

		output := stderr.String()
		if !strings.Contains(output, "Details:") {
			t.Fatalf("expected stderr to contain Details header, got: %q", output)
		}
	})

	t.Run("omits error details when verbose flag is absent", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		exitCode := run([]string{"use", "unsupported-tool", "--no-color"}, &stdout, &stderr)

		if exitCode != 1 {
			t.Fatalf("expected exit code 1, got %d", exitCode)
		}

		output := stderr.String()
		if strings.Contains(output, "Details:") {
			t.Fatalf("expected stderr not to contain Details header, got: %q", output)
		}
	})

	t.Run("returns exit code 0 on successful help execution", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		exitCode := run([]string{"--help"}, &stdout, &stderr)

		if exitCode != 0 {
			t.Fatalf("expected exit code 0, got %d", exitCode)
		}
	})
}
