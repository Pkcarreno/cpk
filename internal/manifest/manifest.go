package manifest

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type PackageSpec struct {
	Name    string
	Version string
	Hash    string
}

func (spec PackageSpec) String() string {
	return spec.Name + "@" + spec.Version + spec.Hash
}

var (
	ErrManifestNotFound      = errors.New("the package.json file does not exist")
	ErrMissingPackageManager = errors.New("the package.json file does not have a packageManager field")
	ErrInvalidPackageSpec    = errors.New("the packageManager field is invalid")
	ErrInvalidManifest       = errors.New("the package.json file has invalid syntax")
)

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

func ParsePackageSpec(raw string) (PackageSpec, error) {
	name, rest, hasAt := strings.Cut(raw, "@")
	if !hasAt || name == "" || rest == "" {
		return PackageSpec{}, ErrInvalidPackageSpec
	}

	version, hash, hasHash := strings.Cut(rest, "+")
	if version == "" {
		return PackageSpec{}, ErrInvalidPackageSpec
	}
	var formattedHash string
	if hasHash {
		formattedHash = "+" + hash
	}

	return PackageSpec{
		Name:    name,
		Version: version,
		Hash:    formattedHash,
	}, nil
}

func ReadPackageSpec(manifestPath string) (PackageSpec, error) {
	manifestBytes, err := Read(manifestPath)
	if err != nil {
		return PackageSpec{}, err
	}

	var payload struct {
		PackageManager *string `json:"packageManager"`
	}

	err = json.Unmarshal(manifestBytes, &payload)
	if err != nil {
		return PackageSpec{}, err
	}

	if payload.PackageManager == nil {
		return PackageSpec{}, ErrMissingPackageManager
	}

	spec, err := ParsePackageSpec(*payload.PackageManager)
	if err != nil {
		return PackageSpec{}, err
	}

	return spec, nil
}

func UpdatePackageSpec(manifestBytes []byte, spec PackageSpec) ([]byte, error) {
	if !json.Valid(manifestBytes) {
		return nil, ErrInvalidManifest
	}
	if !gjson.GetBytes(manifestBytes, "packageManager").Exists() {
		return nil, ErrMissingPackageManager
	}

	return sjson.SetBytes(manifestBytes, "packageManager", spec.String())
}
