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
	ErrIntegrityMismatch    = errors.New("the integrity hash does not match the registry metadata")
	ErrMissingIntegrityHash = errors.New("the manifest does not contain an integrity hash")
)

func NewVerifyCommand() *cobra.Command {
	verifyCommand := &cobra.Command{
		Use:           "verify",
		Short:         "Verify package manager integrity hash against registry",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			packageSpec, err := manifest.ReadPackageSpecAgnostic("./package.json")
			if err != nil {
				return err
			}

			err = registry.ValidatePackageManager(packageSpec.Name)
			if err != nil {
				return err
			}

			if packageSpec.Hash == "" {
				return &ui.ActionableError{
					Summary: "Missing integrity hash",
					Cause:   "The manifest does not contain an integrity hash.",
					Fix:     fmt.Sprintf("Run \"cpk use %s@%s\"", packageSpec.Name, packageSpec.Version),
					Err:     ErrMissingIntegrityHash,
				}
			}

			var customRegistry string
			if registryFlag := cmd.Flags().Lookup("registry"); registryFlag != nil {
				customRegistry = registryFlag.Value.String()
			}

			registryURL := registry.ResolveRegistryURL(customRegistry)

			packageName := resolver.ResolvePackageName(packageSpec.Name, packageSpec.Version)

			uiInstance := ui.FromContext(cmd.Context())
			uiInstance.StartSpinner(fmt.Sprintf("Resolving %s...", packageName))

			metadata, err := registry.FetchPackageMetadata(cmd.Context(), nil, registryURL, packageName)
			if err != nil {
				uiInstance.StopSpinner()
				return err
			}
			uiInstance.StopSpinner()

			packageVersion, hasVersionData := metadata.Versions[packageSpec.Version]
			if !hasVersionData {
				return resolver.ErrVersionNotFound
			}

			integrityHash, err := registry.ConvertIntegrity(packageVersion.Dist)
			if err != nil {
				return err
			}

			if packageSpec.Hash != integrityHash {
				return &ui.ActionableError{
					Summary: "Integrity hash mismatch",
					Cause:   "The integrity hash does not match the published registry metadata.",
					Fix:     fmt.Sprintf("Run \"cpk use %s@%s\"", packageSpec.Name, packageSpec.Version),
					Err:     ErrIntegrityMismatch,
				}
			}

			uiInstance.Success("Integrity hash verified for %s@%s", packageSpec.Name, packageSpec.Version)

			return nil
		},
	}

	return verifyCommand
}
