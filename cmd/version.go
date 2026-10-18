package cmd

import (
	"github.com/pkcarreno/cpk/internal/version"
	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	versionCommand := &cobra.Command{
		Use:           "version",
		Short:         "Display version and build information",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			build := version.Current()
			cmd.Print(build.String())

			return nil
		},
	}

	return versionCommand
}
