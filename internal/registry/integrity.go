package registry

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
)

var (
	ErrInvalidIntegrityHash = errors.New("the integrity hash is invalid")
)

const (
	registryPrefix      = "sha512-"
	integrityHashPrefix = "+sha512."
)

func ConvertIntegrity(rawIntegrity string) (string, error) {
	encodedDigest, hasPrefix := strings.CutPrefix(rawIntegrity, registryPrefix)
	if !hasPrefix {
		return "", ErrInvalidIntegrityHash
	}

	digestBytes, err := base64.StdEncoding.DecodeString(encodedDigest)
	if err != nil {
		return "", ErrInvalidIntegrityHash
	}
	if len(digestBytes) != sha512.Size {
		return "", ErrInvalidIntegrityHash
	}

	return integrityHashPrefix + hex.EncodeToString(digestBytes), nil
}
