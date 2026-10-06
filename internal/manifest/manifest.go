package manifest

import (
	"errors"
	"os"
)

var ErrManifestNotFound = errors.New("the package.json file does not exist")

func Read(manifestPath string) ([]byte, error) {

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrManifestNotFound
		}
		return nil, err
	}

	return data, nil
}
