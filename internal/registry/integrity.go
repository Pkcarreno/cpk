package registry

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
)

var (
	ErrInvalidIntegrityHash = errors.New("the integrity hash is invalid")
	ErrUnsupportedAlgorithm = errors.New("the integrity algorithm is unsupported")
	ErrMissingIntegrity     = errors.New("the package metadata is missing an integrity hash")
)

var supportedSriAlgorithms = map[string]int{
	"sha1":   sha1.Size,
	"sha256": sha256.Size,
	"sha384": sha512.Size384,
	"sha512": sha512.Size,
}

func ConvertIntegrity(dist PackageDist) (string, error) {
	if dist.Integrity != "" {
		algorithm, encodedDigest, hasHyphen := strings.Cut(dist.Integrity, "-")
		if !hasHyphen || algorithm == "" || encodedDigest == "" {
			return "", ErrInvalidIntegrityHash
		}

		expectedDigestSize, isSupported := supportedSriAlgorithms[algorithm]
		if !isSupported {
			return "", ErrUnsupportedAlgorithm
		}

		digestBytes, err := base64.StdEncoding.DecodeString(encodedDigest)
		if err != nil {
			return "", ErrInvalidIntegrityHash
		}
		if len(digestBytes) != expectedDigestSize {
			return "", ErrInvalidIntegrityHash
		}

		return "+" + algorithm + "." + hex.EncodeToString(digestBytes), nil
	}

	if dist.Shasum != "" {
		if len(dist.Shasum) != 40 {
			return "", ErrInvalidIntegrityHash
		}

		for _, character := range dist.Shasum {
			if (character < '0' || character > '9') && (character < 'a' || character > 'f') && (character < 'A' || character > 'F') {
				return "", ErrInvalidIntegrityHash
			}
		}

		return "+sha1." + strings.ToLower(dist.Shasum), nil
	}

	return "", ErrMissingIntegrity
}
