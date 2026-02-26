package igris

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
)

// buildSignedReceipt constructs an ExecutionReceipt with a valid Ed25519
// signature using the canonical-JSON + SHA-256 algorithm expected by
// VerifyReceipt. Returns the receipt and the corresponding public key.
func buildSignedReceipt(t *testing.T, priv ed25519.PrivateKey) *ExecutionReceipt {
	t.Helper()
	r := &ExecutionReceipt{
		ExecutionID:       "0194f3b2-1a2c-7000-8000-000000000001",
		AgentID:           "agent-test",
		TransactionID:     "0194f3b2-1a2c-7000-8000-000000000000",
		TransactionHash:   "sha256:aabbcc",
		CpuTimeMs:         142,
		WallTimeMs:        380,
		MemoryPeakMb:      48,
		FsBytesWritten:    0,
		ToolCalls:         3,
		ViolationOccurred: false,
		TimestampUTC:      "2026-02-21T10:00:00.000Z",
		PreviousHash:      "sha256:001122",
		Hash:              "sha256:placeholder",
		Signature:         "", // filled below
	}

	// Build canonical JSON (exclude signature field, alphabetical key order).
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	delete(m, "signature")
	canonical, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("canonical marshal: %v", err)
	}

	hash := sha256.Sum256(canonical)
	sig := ed25519.Sign(priv, hash[:])
	r.Signature = base64.StdEncoding.EncodeToString(sig)
	return r
}

func TestVerifyReceipt_Valid(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	receipt := buildSignedReceipt(t, priv)

	if err := VerifyReceipt(receipt, pub); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestVerifyReceipt_TamperedField(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	receipt := buildSignedReceipt(t, priv)
	receipt.CpuTimeMs = 9999 // tamper after signing

	if err := VerifyReceipt(receipt, pub); err == nil {
		t.Error("expected error for tampered receipt, got nil")
	}
}

func TestVerifyReceipt_WrongKey(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	wrongPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	receipt := buildSignedReceipt(t, priv)

	if err := VerifyReceipt(receipt, wrongPub); err == nil {
		t.Error("expected error for wrong key, got nil")
	}
}

func TestVerifyReceipt_EmptySignature(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	receipt := &ExecutionReceipt{
		ExecutionID: "test",
		Hash:        "sha256:aabb",
		Signature:   "",
	}
	if err := VerifyReceipt(receipt, pub); err == nil {
		t.Error("expected error for empty signature, got nil")
	}
}

// TestInferResponse_DecodeWithReceipt verifies that a JSON response including
// execution_receipt decodes into InferResponse.ExecutionReceipt correctly.
func TestInferResponse_DecodeWithReceipt(t *testing.T) {
	payload := `{
		"id":"chatcmpl-123",
		"object":"chat.completion",
		"created":1700000000,
		"model":"gpt-4",
		"choices":[{"index":0,"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}],
		"execution_receipt":{
			"execution_id":"0194f3b2-0000-7000-8000-000000000001",
			"cpu_time_ms":142,
			"wall_time_ms":380,
			"memory_peak_mb":48,
			"fs_bytes_written":0,
			"tool_calls":3,
			"violation_occurred":false,
			"hash":"sha256:aabbcc",
			"signature":"dGVzdA=="
		}
	}`

	var resp InferResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.ExecutionReceipt == nil {
		t.Fatal("expected ExecutionReceipt to be set")
	}
	if resp.ExecutionReceipt.ExecutionID != "0194f3b2-0000-7000-8000-000000000001" {
		t.Errorf("unexpected execution_id: %s", resp.ExecutionReceipt.ExecutionID)
	}
	if resp.ExecutionReceipt.CpuTimeMs != 142 {
		t.Errorf("unexpected cpu_time_ms: %d", resp.ExecutionReceipt.CpuTimeMs)
	}
	if resp.ExecutionReceipt.ViolationOccurred {
		t.Error("expected violation_occurred=false")
	}
}

// TestInferResponse_DecodeWithoutReceipt verifies backward compatibility:
// a response without execution_receipt decodes without error.
func TestInferResponse_DecodeWithoutReceipt(t *testing.T) {
	payload := `{
		"id":"chatcmpl-456",
		"object":"chat.completion",
		"created":1700000000,
		"model":"gpt-4",
		"choices":[{"index":0,"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}]
	}`

	var resp InferResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if resp.ExecutionReceipt != nil {
		t.Error("expected ExecutionReceipt to be nil when absent")
	}
}
