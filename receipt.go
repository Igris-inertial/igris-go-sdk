package igris

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// VerifyReceipt verifies the Ed25519 signature of an ExecutionReceipt.
//
// The canonical payload is produced by marshalling all receipt fields except
// "signature" to compact JSON with alphabetically sorted keys (matching
// BTreeMap serialisation used by the Runtime). The payload is SHA-256 hashed
// and the signature is verified against the hash using the provided public key.
//
// Returns nil on success. Returns a non-nil error if the signature is missing,
// malformed, or does not verify.
//
// Call this explicitly after receiving a response that includes ExecutionReceipt.
// Verification is never performed automatically. The public key must match
// IGRIS_RUNTIME_PUBLIC_KEY configured on the Runtime.
func VerifyReceipt(receipt *ExecutionReceipt, pub ed25519.PublicKey) error {
	if receipt == nil {
		return fmt.Errorf("receipt: nil receipt")
	}
	if len(pub) != ed25519.PublicKeySize {
		return fmt.Errorf("receipt: invalid public key length %d", len(pub))
	}

	sigBytes, err := base64.StdEncoding.DecodeString(receipt.Signature)
	if err != nil {
		return fmt.Errorf("receipt: signature base64 decode: %w", err)
	}
	if len(sigBytes) == 0 {
		return fmt.Errorf("receipt: empty signature")
	}

	// Marshal receipt to JSON then back to a map so that json.Marshal produces
	// alphabetically sorted keys — matching the BTreeMap canonical form used
	// by the Runtime and Overture's verifySignedJSON.
	raw, err := json.Marshal(receipt)
	if err != nil {
		return fmt.Errorf("receipt: marshal: %w", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("receipt: unmarshal: %w", err)
	}
	delete(m, "signature")

	canonical, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("receipt: canonical marshal: %w", err)
	}

	hash := sha256.Sum256(canonical)
	if !ed25519.Verify(pub, hash[:], sigBytes) {
		return fmt.Errorf("receipt: signature verification failed")
	}
	return nil
}
