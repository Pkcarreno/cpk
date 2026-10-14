package cmd

import (
	"errors"
	"fmt"

	"github.com/pkcarreno/cpk/internal/manifest"
	"github.com/pkcarreno/cpk/internal/registry"
	"github.com/pkcarreno/cpk/internal/resolver"
	"github.com/pkcarreno/cpk/internal/ui"
	"github.com/spf13/cobra"
)

var (
	ErrIntegrityMismatch = errors.New("the integrity hash does not match the registry metadata")
	ErrOutdatedPackage   = errors.New("the package manager is outdated")
)

func NewCheckCommand() *cobra.Command {
	checkCommand := &cobra.Command{
		Use:   "check",
		Short: "Verify package manager updates and integrity",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			packageSpec, err := manifest.ReadPackageSpec("./package.json")
			if err != nil {
				return err
			}

			err = registry.ValidatePackageManager(packageSpec.Name)
			if err != nil {
				return err
			}

			var customRegistry string
			if registryFlag := cmd.Flags().Lookup("registry"); registryFlag != nil {
				customRegistry = registryFlag.Value.String()
			}

			registryURL := registry.ResolveRegistryURL(customRegistry)

			targetVersion := "latest"
			if packageSpec.Name == "yarn" && resolver.ResolvePackageName(packageSpec.Name, packageSpec.Version) == "yarn" {
				targetVersion = "1"
			}

			packageName := resolver.ResolvePackageName(packageSpec.Name, targetVersion)

			uiInstance := ui.FromContext(cmd.Context())
			uiInstance.StartSpinner(fmt.Sprintf("Resolving %s...", packageName))

			metadata, err := registry.FetchPackageMetadata(cmd.Context(), nil, registryURL, packageName)
			if err != nil {
				uiInstance.StopSpinner()
				return err
			}
			uiInstance.StopSpinner()

			resolvedVersion, err := resolver.ResolveVersion(metadata, targetVersion)
			if err != nil {
				return err
			}

			packageVersion, hasVersionData := metadata.Versions[packageSpec.Version]
			if !hasVersionData {
				return resolver.ErrVersionNotFound
			}

			integrityHash, err := registry.ConvertIntegrity(packageVersion.Dist)
			if err != nil {
				return err
			}

			if packageSpec.Hash != integrityHash {
				return ErrIntegrityMismatch
			}

			isOutdated, err := resolver.IsLowerVersion(packageSpec.Version, resolvedVersion)
			if err != nil {
				return err
			}

			if isOutdated {
				uiInstance.Warn("Update available for %s: %s -> %s", packageSpec.Name, packageSpec.Version, resolvedVersion)
				if isFailOnOutdatedFlag, _ := cmd.Flags().GetBool("fail-on-outdated"); isFailOnOutdatedFlag {
					return ErrOutdatedPackage
				}
			} else {
				uiInstance.Info("%s is already up to date (%s)", packageSpec.Name, packageSpec.Version)
			}

			return nil
		},
	}

	checkCommand.Flags().Bool("fail-on-outdated", false, "Exit with a non-zero code if the package manager is outdated")

	return checkCommand
}
