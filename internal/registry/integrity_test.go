package registry_test

import (
	"errors"
	"testing"

	"github.com/pkcarreno/cpk/internal/registry"
)

func TestConvertIntegrity(t *testing.T) {
	t.Run("returns formatted hash when the integrity string is valid", func(t *testing.T) {
		rawIntegrity := "sha512-hFbAm884LpT8E1I+sI2hcVojcCcUf+ylT5Yr34U3qRfH//2nAt7G1NFYmqGTDpRPnL1xZnMxIrA9/xuatbuZXQ=="
		expected := "+sha512.8456c09bcf382e94fc13523eb08da1715a237027147feca54f962bdf8537a917c7fffda702dec6d4d1589aa1930e944f9cbd7166733122b03dff1b9ab5bb995d"

		actualHash, err := registry.ConvertIntegrity(rawIntegrity)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if actualHash != expected {
			t.Fatalf("expected: %q, got: %q", expected, actualHash)
		}
	})

	t.Run("returns ErrInvalidIntegrityHash when the algorithm is not sha512", func(t *testing.T) {
		rawIntegrity := "sha256-AAAA"

		_, err := registry.ConvertIntegrity(rawIntegrity)

		if !errors.Is(err, registry.ErrInvalidIntegrityHash) {
			t.Fatalf("expected ErrInvalidIntegrityHash, got: %v", err)
		}
	})

	t.Run("returns ErrInvalidIntegrityHash when the base64 is malformed", func(t *testing.T) {
		rawIntegrity := "sha512-invalid%base64"

		_, err := registry.ConvertIntegrity(rawIntegrity)

		if !errors.Is(err, registry.ErrInvalidIntegrityHash) {
			t.Fatalf("expected ErrInvalidIntegrityHash, got: %v", err)
		}
	})

	t.Run("returns ErrInvalidIntegrityHash when the decoded length is not 64 bytes", func(t *testing.T) {
		rawIntegrity := "sha512-hFbAm884LpT8E1I+sI2hcVojcCcUf+ylT5Yr34U3qRfH"

		_, err := registry.ConvertIntegrity(rawIntegrity)

		if !errors.Is(err, registry.ErrInvalidIntegrityHash) {
			t.Fatalf("expected ErrInvalidIntegrityHash, got: %v", err)
		}
	})

	t.Run("returns ErrInvalidIntegrityHash when the input is empty", func(t *testing.T) {
		rawIntegrity := ""

		_, err := registry.ConvertIntegrity(rawIntegrity)

		if !errors.Is(err, registry.ErrInvalidIntegrityHash) {
			t.Fatalf("expected ErrInvalidIntegrityHash, got: %v", err)
		}
	})
}
