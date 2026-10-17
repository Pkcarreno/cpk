package ui

import (
	"errors"
	"strings"

	"github.com/pkcarreno/cpk/internal/manifest"
	"github.com/pkcarreno/cpk/internal/registry"
	"github.com/pkcarreno/cpk/internal/resolver"
)

type ActionableError struct {
	Summary string
	Cause   string
	Fix     string
	Err     error
}

func (err *ActionableError) Error() string {
	return err.Summary
}

func (err *ActionableError) Unwrap() error {
	return err.Err
}

func FormatError(err error, isVerbose bool, noColor bool) string {
	if actionableError, isActionable := toActionableError(err); isActionable {
		var builder strings.Builder

		if noColor {
			builder.WriteString("Error: " + actionableError.Summary + "\n")
			builder.WriteString("Cause: " + actionableError.Cause + "\n")
			builder.WriteString("Fix:   " + actionableError.Fix + "\n")
		} else {
			builder.WriteString(ansiRed + "Error:" + ansiReset + " " + actionableError.Summary + "\n")
			builder.WriteString(ansiYellow + "Cause:" + ansiReset + " " + actionableError.Cause + "\n")
			builder.WriteString(ansiCyan + "Fix:" + ansiReset + "   " + actionableError.Fix + "\n")
		}

		if isVerbose && actionableError.Err != nil {
			builder.WriteString("\nDetails:\n" + actionableError.Err.Error() + "\n")
		}

		return builder.String()
	}

	output := "Error: " + err.Error() + "\n"
	if !noColor {
		output = ansiRed + "Error: " + ansiReset + " " + err.Error() + "\n"
	}
	if isVerbose {
		output += "\nDetails:\n" + err.Error() + "\n"
	}
	return output
}

func toActionableError(err error) (*ActionableError, bool) {
	var actionableError *ActionableError
	if errors.As(err, &actionableError) {
		return actionableError, true
	}

	switch {
	case errors.Is(err, manifest.ErrManifestNotFound):
		return &ActionableError{
			Summary: "package.json file not found",
			Cause:   "The package.json file does not exist in the current directory.",
			Fix:     "Run \"npm init -y\" to create a manifest, or go to a Node.js project directory.",
			Err:     err,
		}, true

	case errors.Is(err, manifest.ErrMissingPackageManager):
		return &ActionableError{
			Summary: "Missing packageManager field",
			Cause:   "The package.json file does not contain a packageManager field.",
			Fix:     "Run \"cpk use <package-manager>\" to set a package manager.",
			Err:     err,
		}, true

	case errors.Is(err, manifest.ErrInvalidPackageSpec):
		return &ActionableError{
			Summary: "Invalid package specification",
			Cause:   "The packageManager field does not match the format name@version[+hash].",
			Fix:     "Run \"cpk use <package-manager>@<version>\" to set a valid package specification.",
			Err:     err,
		}, true

	case errors.Is(err, manifest.ErrInvalidManifest):
		return &ActionableError{
			Summary: "Invalid package.json manifest",
			Cause:   "The package.json file contains invalid JSON syntax.",
			Fix:     "Correct the JSON syntax errors in the package.json file.",
			Err:     err,
		}, true

	case errors.Is(err, manifest.ErrPackageManagerExists):
		return &ActionableError{
			Summary: "packageManager field already exists",
			Cause:   "The package.json file already contains a packageManager field.",
			Fix:     "Run \"cpk use <package-manager>\" to update the existing packageManager field.",
			Err:     err,
		}, true

	case errors.Is(err, manifest.ErrInvalidIntegrityHash) || errors.Is(err, registry.ErrInvalidIntegrityHash):
		return &ActionableError{
			Summary: "Invalid integrity hash",
			Cause:   "The integrity hash format or digest length does not match supported algorithms.",
			Fix:     "Run \"cpk use <package-manager>\" to get a fresh hash from the registry.",
			Err:     err,
		}, true

	case errors.Is(err, registry.ErrUnsupportedPackageManager):
		return &ActionableError{
			Summary: "Unsupported package manager",
			Cause:   "cpk only supports pnpm, yarn, and npm.",
			Fix:     "Run \"cpk use <pnpm|yarn|npm>\" instead.",
			Err:     err,
		}, true

	case errors.Is(err, registry.ErrUnsupportedAlgorithm):
		return &ActionableError{
			Summary: "Unsupported integrity algorithm",
			Cause:   "The registry integrity hash uses an unsupported algorithm.",
			Fix:     "Use a version with sha512, sha384, sha256, sha224, or sha1 hashes.",
			Err:     err,
		}, true

	case errors.Is(err, registry.ErrMissingIntegrity):
		return &ActionableError{
			Summary: "Missing integrity hash in registry",
			Cause:   "The package metadata does not contain an integrity hash or shasum.",
			Fix:     "Select a package version that includes published integrity metadata.",
			Err:     err,
		}, true

	case errors.Is(err, registry.ErrPackageNotFound):
		return &ActionableError{
			Summary: "Package not found in registry",
			Cause:   "The package does not exist in the configured npm registry.",
			Fix:     "Make sure the package name and registry URL are correct.",
			Err:     err,
		}, true

	case errors.Is(err, registry.ErrRegistryRequestFailed):
		return &ActionableError{
			Summary: "Registry request failed",
			Cause:   "The registry returned an error or the network request failed.",
			Fix:     "Check your network connection and registry URL.",
			Err:     err,
		}, true

	case errors.Is(err, resolver.ErrVersionNotFound):
		return &ActionableError{
			Summary: "Version not found in registry",
			Cause:   "The registry does not have a release matching the target version.",
			Fix:     "Examine the package versions on npmjs.com.",
			Err:     err,
		}, true

	case strings.Contains(err.Error(), "the integrity hash does not match"):
		return &ActionableError{
			Summary: "Integrity hash mismatch",
			Cause:   "The integrity hash does not match the published registry metadata.",
			Fix:     "Run \"cpk use <package-manager>@<version>\" to update the manifest with the valid hash.",
			Err:     err,
		}, true

	case strings.Contains(err.Error(), "the package manager is outdated"):
		return &ActionableError{
			Summary: "Package manager is outdated",
			Cause:   "A newer version of the package manager exists in the registry.",
			Fix:     "Run \"cpk up\" to update to the latest version.",
			Err:     err,
		}, true

	case strings.Contains(err.Error(), "cannot downgrade with up"):
		return &ActionableError{
			Summary: "Downgrade rejected by up command",
			Cause:   "The target version is lower than the active pinned version.",
			Fix:     "Run \"cpk use <package-manager>@<version>\" to force a downgrade.",
			Err:     err,
		}, true

	default:
		return nil, false
	}
}
