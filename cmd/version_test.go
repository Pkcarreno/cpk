package cmd_test

import (
	"bytes"
	"testing"

	"github.com/pkcarreno/cpk/cmd"
	"github.com/pkcarreno/cpk/internal/version"
)

func TestNewVersionCommand(t *testing.T) {
	t.Run("outputs multiline diagnostic metadata to stdout", func(t *testing.T) {
		versionCommand := cmd.NewVersionCommand()

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		versionCommand.SetOut(&stdout)
		versionCommand.SetErr(&stderr)

		err := versionCommand.Execute()
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}

		if stderr.Len() != 0 {
			t.Fatalf("expected stderr to be empty, got: %q", stderr.String())
		}

		if stdout.String() != version.Current().String() {
			t.Fatalf("stdout does not match expected output, expected: %q, got: %q", version.Current().String(), stdout.String())
		}
	})
}
