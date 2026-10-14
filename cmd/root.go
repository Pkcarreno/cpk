package cmd

import (
	"github.com/pkcarreno/cpk/internal/ui"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "cpk",
		Short: "Manage packageManager key in the package.json file",
		Long: "cpk manages the packageManager field in package.json.\n\n" +
			"It resolves package manager versions and SHA-512 integrity hashes from\n" +
			"the npm registry. It updates the package.json file directly without\n" +
			"changing the existing key order, indentation, or formatting.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(command *cobra.Command, arguments []string) error {
			commandFlags := command.Flags()

			hasQuietFlag, _ := commandFlags.GetBool("quiet")
			hasVerboseFlag, _ := commandFlags.GetBool("verbose")
			hasNoColorFlag, _ := commandFlags.GetBool("no-color")

			isTerminal := ui.IsTerminal(command.ErrOrStderr()) && !ui.IsCI()
			hasNoColor := hasNoColorFlag || ui.HasNoColor()

			uiInstance := ui.New(ui.Config{
				Out:       command.OutOrStdout(),
				Err:       command.ErrOrStderr(),
				IsTTY:     isTerminal,
				IsQuiet:   hasQuietFlag,
				IsVerbose: hasVerboseFlag,
				NoColor:   hasNoColor,
			})

			command.SetContext(ui.WithContext(command.Context(), uiInstance))
			return nil
		},
	}

	rootCommand.PersistentFlags().String("registry", "", "URL of the npm registry")
	rootCommand.PersistentFlags().BoolP("quiet", "q", false, "Do not write status messages to stderr")
	rootCommand.PersistentFlags().BoolP("verbose", "v", false, "Write detailed execution steps to stderr")
	rootCommand.PersistentFlags().Bool("no-color", false, "Do not use ANSI color codes")

	rootCommand.AddCommand(NewCurrentCommand())
	rootCommand.AddCommand(NewUseCommand())
	rootCommand.AddCommand(NewUpCommand())
	rootCommand.AddCommand(NewCheckCommand())

	return rootCommand
}

func Execute() error {
	return NewRootCommand().Execute()
}
