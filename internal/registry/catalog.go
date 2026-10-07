package registry

import (
	"errors"
)

var (
	ErrUnsupportedPackageManager = errors.New("the package manager is not supported")
)

func ValidatePackageManager(packageManager string) error {
	switch packageManager {
	case "pnpm", "yarn", "npm":
		return nil
	default:
		return ErrUnsupportedPackageManager
	}

}
