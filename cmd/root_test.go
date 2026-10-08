package cmd_test

import (
	"testing"

	"github.com/pkcarreno/cpk/cmd"
)

func TestNewRootCommand(t *testing.T) {
	t.Run("configures the root command name", func(t *testing.T) {
		rootCommand := cmd.NewRootCommand()
		if rootCommand.Use != "cpk" {
			t.Fatalf("expected command use 'cpk', got %q", rootCommand.Use)
		}
	})

	t.Run("binds the persistent registry flag with empty default", func(t *testing.T) {
		rootCommand := cmd.NewRootCommand()
		flag := rootCommand.PersistentFlags().Lookup("registry")
		if flag == nil {
			t.Fatal("the persistent flag --registry is not defined")
		}
		if flag.DefValue != "" {
			t.Fatalf("expected empty default value for --registry, got %q", flag.DefValue)
		}
	})

	t.Run("parses a custom registry flag value", func(t *testing.T) {
		rootCommand := cmd.NewRootCommand()
		testURL := "https://npm.example.com"

		if err := rootCommand.ParseFlags([]string{"--registry", testURL}); err != nil {
			t.Fatalf("unexpected flag parse error: %v", err)
		}

		flag := rootCommand.PersistentFlags().Lookup("registry")
		if flag.Value.String() != testURL {
			t.Fatalf("expected flag value %q, got %q", testURL, flag.Value.String())
		}
	})
}

func TestExecute(t *testing.T) {
	t.Run("execute the root command", func(t *testing.T) {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected execution error %v", err)
		}
	})
}
