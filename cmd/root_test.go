package cmd_test

import (
	"context"
	"testing"

	"github.com/pkcarreno/cpk/cmd"
	"github.com/pkcarreno/cpk/internal/ui"
	"github.com/spf13/cobra"
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

		if parseError := rootCommand.ParseFlags([]string{"--registry", testURL}); parseError != nil {
			t.Fatalf("unexpected flag parse error: %v", parseError)
		}

		flag := rootCommand.PersistentFlags().Lookup("registry")
		if flag.Value.String() != testURL {
			t.Fatalf("expected flag value %q, got %q", testURL, flag.Value.String())
		}
	})

	t.Run("binds persistent flags for output control", func(t *testing.T) {
		rootCommand := cmd.NewRootCommand()

		expectedFlags := []struct {
			name      string
			shorthand string
			defValue  string
		}{
			{name: "quiet", shorthand: "q", defValue: "false"},
			{name: "verbose", shorthand: "v", defValue: "false"},
			{name: "no-color", shorthand: "", defValue: "false"},
		}

		for _, expected := range expectedFlags {
			flag := rootCommand.PersistentFlags().Lookup(expected.name)
			if flag == nil {
				t.Fatalf("expected persistent flag --%s to be defined", expected.name)
			}
			if flag.Shorthand != expected.shorthand {
				t.Fatalf("expected shorthand for --%s to be %q, got %q", expected.name, expected.shorthand, flag.Shorthand)
			}
			if flag.DefValue != expected.defValue {
				t.Fatalf("expected default value for --%s to be %q, got %q", expected.name, expected.defValue, flag.DefValue)
			}
		}
	})

	t.Run("parses output control flags", func(t *testing.T) {
		rootCommand := cmd.NewRootCommand()

		if parseError := rootCommand.ParseFlags([]string{"-q", "-v", "--no-color"}); parseError != nil {
			t.Fatalf("unexpected flag parse error: %v", parseError)
		}

		for _, flagName := range []string{"quiet", "verbose", "no-color"} {
			flag := rootCommand.PersistentFlags().Lookup(flagName)
			if flag.Value.String() != "true" {
				t.Fatalf("expected flag --%s to be true, got %q", flagName, flag.Value.String())
			}
		}
	})

	t.Run("stores configured ui instance in command context", func(t *testing.T) {
		rootCommand := cmd.NewRootCommand()

		var recordedContext context.Context
		dummyCommand := &cobra.Command{
			Use: "dummy",
			Run: func(command *cobra.Command, arguments []string) {
				recordedContext = command.Context()
			},
		}
		rootCommand.AddCommand(dummyCommand)
		rootCommand.SetArgs([]string{"dummy", "--quiet"})

		if executionError := rootCommand.Execute(); executionError != nil {
			t.Fatalf("unexpected execution error: %v", executionError)
		}

		if recordedContext == nil {
			t.Fatal("expected command context to be recorded")
		}

		uiInstance := ui.FromContext(recordedContext)
		if uiInstance == nil {
			t.Fatal("expected ui instance in context")
		}
	})
}
