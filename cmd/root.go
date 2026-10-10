package cmd

import (
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
	}

	rootCommand.PersistentFlags().String("registry", "", "URL of the npm registry")

	rootCommand.AddCommand(NewCurrentCommand())
	rootCommand.AddCommand(NewUseCommand())

	return rootCommand
}

func Execute() error {
	return NewRootCommand().Execute()
}
