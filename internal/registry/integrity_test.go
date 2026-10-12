package registry_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pkcarreno/cpk/internal/registry"
)

func TestConvertIntegrity(t *testing.T) {
	validCases := []struct {
		name         string
		dist         registry.PackageDist
		expectedHash string
	}{
		{
			name: "converts sha512 integrity to lowercase hexadecimal format",
			dist: registry.PackageDist{
				Integrity: "sha512-hFbAm884LpT8E1I+sI2hcVojcCcUf+ylT5Yr34U3qRfH//2nAt7G1NFYmqGTDpRPnL1xZnMxIrA9/xuatbuZXQ==",
			},
			expectedHash: "+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d",
		},
		{
			name: "converts sha384 integrity to lowercase hexadecimal format",
			dist: registry.PackageDist{
				Integrity: "sha384-/b2OdaZ/KfcBpOBAOF4uI5hjA+oQI5IRr5B/y7g1eLPkF8txzmRu/QgZ3YwIjeG9",
			},
			expectedHash: "+sha384.fdbd8e75a67f29f701a4e040385e2e23986303ea10239211af907fcbb83578b3e417cb71ce646efd0819dd8c088de1bd",
		},
		{
			name: "converts sha256 integrity to lowercase hexadecimal format",
			dist: registry.PackageDist{
				Integrity: "sha256-uU0nuZNNPgilLlLX2n2r+sSE7+N6U4DukIj3rOLvzek=",
			},
			expectedHash: "+sha256.b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name: "converts sha1 integrity to lowercase hexadecimal format",
			dist: registry.PackageDist{
				Integrity: "sha1-Kq5sNclPz7QV2+lfQIuc6R7oRu0=",
			},
			expectedHash: "+sha1.2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
		},
		{
			name: "converts legacy shasum to sha1 hash when integrity is absent",
			dist: registry.PackageDist{
				Shasum: "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
			},
			expectedHash: "+sha1.2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
		},
		{
			name: "converts uppercase legacy shasum to lowercase sha1 hash",
			dist: registry.PackageDist{
				Shasum: "2AAE6C35C94FCFB415DBE95F408B9CE91EE846ED",
			},
			expectedHash: "+sha1.2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
		},
		{
			name: "prefers integrity over shasum when both are present",
			dist: registry.PackageDist{
				Integrity: "sha512-hFbAm884LpT8E1I+sI2hcVojcCcUf+ylT5Yr34U3qRfH//2nAt7G1NFYmqGTDpRPnL1xZnMxIrA9/xuatbuZXQ==",
				Shasum:    "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
			},
			expectedHash: "+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d",
		},
	}

	for _, testCase := range validCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualHash, err := registry.ConvertIntegrity(testCase.dist)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actualHash != testCase.expectedHash {
				t.Fatalf("expected: %q, got: %q", testCase.expectedHash, actualHash)
			}
		})
	}

	invalidCases := []struct {
		name        string
		dist        registry.PackageDist
		expectedErr error
	}{
		{
			name: "returns ErrMissingIntegrity when integrity and shasum are both missing",
			dist: registry.PackageDist{
				Integrity: "",
				Shasum:    "",
			},
			expectedErr: registry.ErrMissingIntegrity,
		},
		{
			name: "returns ErrUnsupportedAlgorithm when integrity algorithm is unsupported",
			dist: registry.PackageDist{
				Integrity: "md5-1B2M2Y8AsgTpgAmY7PhCfg==",
			},
			expectedErr: registry.ErrUnsupportedAlgorithm,
		},
		{
			name: "returns ErrUnsupportedAlgorithm when algorithm is sha224",
			dist: registry.PackageDist{
				Integrity: "sha224-AAAA",
			},
			expectedErr: registry.ErrUnsupportedAlgorithm,
		},
		{
			name: "returns ErrInvalidIntegrityHash when base64 payload is malformed",
			dist: registry.PackageDist{
				Integrity: "sha512-invalid%base64",
			},
			expectedErr: registry.ErrInvalidIntegrityHash,
		},
		{
			name: "returns ErrInvalidIntegrityHash when decoded digest length is incorrect",
			dist: registry.PackageDist{
				Integrity: "sha512-hFbAm884LpT8E1I+sI2hcVojcCcUf+ylT5Yr34U3qRfH",
			},
			expectedErr: registry.ErrInvalidIntegrityHash,
		},
		{
			name: "returns ErrInvalidIntegrityHash when legacy shasum length is short",
			dist: registry.PackageDist{
				Shasum: strings.Repeat("a", 39),
			},
			expectedErr: registry.ErrInvalidIntegrityHash,
		},
		{
			name: "returns ErrInvalidIntegrityHash when legacy shasum length is long",
			dist: registry.PackageDist{
				Shasum: strings.Repeat("a", 41),
			},
			expectedErr: registry.ErrInvalidIntegrityHash,
		},
		{
			name: "returns ErrInvalidIntegrityHash when legacy shasum contains non-hexadecimal characters",
			dist: registry.PackageDist{
				Shasum: strings.Repeat("z", 40),
			},
			expectedErr: registry.ErrInvalidIntegrityHash,
		},
	}

	for _, testCase := range invalidCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := registry.ConvertIntegrity(testCase.dist)
			if !errors.Is(err, testCase.expectedErr) {
				t.Fatalf("expected error %v, got: %v", testCase.expectedErr, err)
			}
		})
	}
}
