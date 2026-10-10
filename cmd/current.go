package cmd

import (
	"github.com/pkcarreno/cpk/internal/manifest"
	"github.com/pkcarreno/cpk/internal/registry"
	"github.com/pkcarreno/cpk/internal/resolver"
	"github.com/spf13/cobra"
)

func NewCurrentCommand() *cobra.Command {
	currentCommand := &cobra.Command{
		Use:   "current",
		Short: "Display the active package manager and version",
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

			hasCheckFlag, _ := cmd.Flags().GetBool("check")
			if !hasCheckFlag {
				return nil
			}

			if err := registry.ValidatePackageManager(packageSpec.Name); err != nil {
				return err
			}

			var customRegistry string
			if registryFlag := cmd.Flags().Lookup("registry"); registryFlag != nil {
				customRegistry = registryFlag.Value.String()
			}

			registryURL := registry.ResolveRegistryURL(customRegistry)

			metadata, err := registry.FetchPackageMetadata(cmd.Context(), nil, registryURL, packageSpec.Name)
			if err != nil {
				return err
			}

			latestVersion, err := resolver.ResolveVersion(metadata, "latest")
			if err != nil {
				return err
			}

			if packageSpec.Version != latestVersion {
				cmd.Printf("\nUpdate available: %s -> %s\n", packageSpec.Version, latestVersion)
			} else {
				cmd.Printf("\nNo updates available\n")
			}

			return nil
		},
	}

	currentCommand.Flags().Bool("raw", false, "display only the packageManager string without labels")
	currentCommand.Flags().Bool("check", false, "query the registry for newer versions")

	return currentCommand
}
