package cmd

import (
	"strings"

	"github.com/pkcarreno/cpk/internal/manifest"
	"github.com/pkcarreno/cpk/internal/registry"
	"github.com/pkcarreno/cpk/internal/resolver"
	"github.com/spf13/cobra"
)

func NewUseCommand() *cobra.Command {
	useCommand := &cobra.Command{
		Use:   "use <package-manager[@target]>",
		Short: "Set the package manager and version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := manifest.Read("./package.json"); err != nil {
				return err
			}

			packageManager, targetVersion, hasTarget := strings.Cut(args[0], "@")
			if !hasTarget {
				targetVersion = "latest"
			}
			if hasTarget && targetVersion == "" {
				return resolver.ErrVersionNotFound
			}

			err := registry.ValidatePackageManager(packageManager)
			if err != nil {
				return err
			}

			var customRegistry string
			if registryFlag := cmd.Flags().Lookup("registry"); registryFlag != nil {
				customRegistry = registryFlag.Value.String()
			}

			registryURL := registry.ResolveRegistryURL(customRegistry)

			packageName := resolver.ResolvePackageName(packageManager, targetVersion)

			metadata, err := registry.FetchPackageMetadata(cmd.Context(), nil, registryURL, packageName)
			if err != nil {
				return err
			}

			resolvedVersion, err := resolver.ResolveVersion(metadata, targetVersion)
			if err != nil {
				return err
			}

			packageVersion, hasVersionData := metadata.Versions[resolvedVersion]
			if !hasVersionData {
				return resolver.ErrVersionNotFound
			}

			integrityHash, err := registry.ConvertIntegrity(packageVersion.Dist)
			if err != nil {
				return err
			}

			targetSpec := manifest.PackageSpec{
				Name:    packageManager,
				Version: resolvedVersion,
				Hash:    integrityHash,
			}

			_, err = manifest.WritePackageSpec("./package.json", targetSpec)
			if err != nil {
				return err
			}

			return nil
		},
	}

	return useCommand
}
