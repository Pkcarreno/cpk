package manifest

import (
	"bytes"
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
	ErrPackageManagerExists  = errors.New("the package.json file already has a packageManager field")
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
	if version == "" || (hasHash && hash == "") {
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

	if !json.Valid(manifestBytes) {
		return PackageSpec{}, ErrInvalidManifest
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

func InsertPackageSpec(manifestBytes []byte, spec PackageSpec) ([]byte, error) {
	if !json.Valid(manifestBytes) {
		return nil, ErrInvalidManifest
	}
	if gjson.GetBytes(manifestBytes, "packageManager").Exists() {
		return nil, ErrPackageManagerExists
	}

	closeBraceIndex := bytes.LastIndexByte(manifestBytes, '}')
	if closeBraceIndex == -1 {
		return nil, ErrInvalidManifest
	}

	trimmedPrefixBytes := bytes.TrimRight(manifestBytes[:closeBraceIndex], " \t\r\n")
	hasKeys := len(trimmedPrefixBytes) > 0 && trimmedPrefixBytes[len(trimmedPrefixBytes)-1] != '{'

	indent := detectIndentation(manifestBytes)
	var buffer bytes.Buffer
	buffer.Write(trimmedPrefixBytes)

	if hasKeys {
		buffer.WriteString(",")
	}
	if indent != "" {
		buffer.WriteString("\n")
		buffer.WriteString(indent)
		buffer.WriteString(`"packageManager": "` + spec.String() + `"`)
		buffer.WriteString("\n")
	} else {
		buffer.WriteString(`"packageManager":"` + spec.String() + `"`)
	}

	buffer.Write(manifestBytes[closeBraceIndex:])
	return buffer.Bytes(), nil
}

func WritePackageSpec(manifestPath string, targetSpec PackageSpec) (isModified bool, err error) {
	manifestBytes, err := Read(manifestPath)
	if err != nil {
		return false, err
	}

	if !json.Valid(manifestBytes) {
		return false, ErrInvalidManifest
	}

	currentValue := gjson.GetBytes(manifestBytes, "packageManager")

	var updatedBytes []byte

	if currentValue.Exists() {
		if currentValue.String() == targetSpec.String() {
			return false, nil
		}

		updatedBytes, err = UpdatePackageSpec(manifestBytes, targetSpec)
		if err != nil {
			return false, err
		}
	} else {
		updatedBytes, err = InsertPackageSpec(manifestBytes, targetSpec)
		if err != nil {
			return false, err
		}
	}

	info, err := os.Stat(manifestPath)
	if err != nil {
		return false, err
	}

	err = os.WriteFile(manifestPath, updatedBytes, info.Mode().Perm())
	if err != nil {
		return false, err
	}

	return true, nil
}

func detectIndentation(manifestBytes []byte) string {
	lines := bytes.Split(manifestBytes, []byte("\n"))
	for _, line := range lines {
		trimmed := bytes.TrimLeft(line, " \t")
		if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("{")) || bytes.Equal(trimmed, []byte("}")) {
			continue
		}
		indent := line[:len(line)-len(trimmed)]
		if len(indent) > 0 {
			if indent[0] == '\t' {
				return "\t"
			}
			if len(indent) >= 4 && bytes.Equal(indent[:4], []byte("    ")) {
				return "    "
			}
			if len(indent) >= 2 && bytes.Equal(indent[:2], []byte("  ")) {
				return "  "
			}
		}
	}
	return ""
}
