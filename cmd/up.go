package cmd

import (
	"errors"
	"fmt"

	"github.com/pkcarreno/cpk/internal/manifest"
	"github.com/pkcarreno/cpk/internal/registry"
	"github.com/pkcarreno/cpk/internal/resolver"
	"github.com/spf13/cobra"
)

var (
	ErrDowngrade = errors.New("cannot downgrade with up: run 'cpk use' instead")
)

func NewUpCommand() *cobra.Command {
	upCommand := &cobra.Command{
		Use:   "up [target]",
		Short: "Update the package manager version",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageSpec, err := manifest.ReadPackageSpec("./package.json")
			if err != nil {
				if errors.Is(err, manifest.ErrMissingPackageManager) {
					return fmt.Errorf("%w: run 'cpk use <package-manager>' first", err)
				}
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

			metadata, err := registry.FetchPackageMetadata(cmd.Context(), nil, registryURL, packageSpec.Name)
			if err != nil {
				return err
			}

			targetVersion := "latest"
			if len(args) > 0 {
				if args[0] == "" {
					return resolver.ErrVersionNotFound
				}
				targetVersion = args[0]
			}

			resolvedVersion, err := resolver.ResolveVersion(metadata, targetVersion)
			if err != nil {
				return err
			}

			isDowngrade, err := resolver.IsLowerVersion(resolvedVersion, packageSpec.Version)
			if err != nil {
				return err
			}

			if isDowngrade {
				return fmt.Errorf("%w: target version %s is lower than the current version %s", ErrDowngrade, resolvedVersion, packageSpec.Version)
			}

			packageVersion, hasVersion := metadata.Versions[resolvedVersion]
			if !hasVersion {
				return resolver.ErrVersionNotFound
			}

			integrityHash, err := registry.ConvertIntegrity(packageVersion.Dist)
			if err != nil {
				return err
			}

			targetSpec := manifest.PackageSpec{
				Name:    packageSpec.Name,
				Version: resolvedVersion,
				Hash:    integrityHash,
			}

			isModified, err := manifest.WritePackageSpec("./package.json", targetSpec)
			if err != nil {
				return err
			}

			if isModified {
				cmd.Printf("Updated %s from %s to %s\n", packageSpec.Name, packageSpec.Version, resolvedVersion)
			} else {
				cmd.Printf("%s is already up to date (%s)\n", packageSpec.Name, packageSpec.Version)
			}

			return nil
		},
	}

	return upCommand
}
