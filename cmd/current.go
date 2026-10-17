package cmd

import (
	"github.com/pkcarreno/cpk/internal/manifest"
	"github.com/spf13/cobra"
)

func NewCurrentCommand() *cobra.Command {
	currentCommand := &cobra.Command{
		Use:           "current",
		Short:         "Display the active package manager and version",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			packageSpec, err := manifest.ReadPackageSpec("./package.json")
			if err != nil {
				return err
			}

			if isRawFlag, _ := cmd.Flags().GetBool("raw"); isRawFlag {
				cmd.Println(packageSpec.String())
				return nil
			}

			cmd.Printf("Package Manager:\t%s\nVersion:        \t%s\nIntegrity Hash: \t%s\n", packageSpec.Name, packageSpec.Version, packageSpec.Hash)

			return nil
		},
	}

	currentCommand.Flags().Bool("raw", false, "display only the packageManager string without labels")

	return currentCommand
}
