package igris

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// -- Bounds tests ---------------------------------------------------------

func TestBoundsDefault(t *testing.T) {
	b := DefaultBounds()
	if b.CpuPercent != 80 {
		t.Errorf("expected CpuPercent 80, got %d", b.CpuPercent)
	}
	if b.MemoryMb != 512 {
		t.Errorf("expected MemoryMb 512, got %d", b.MemoryMb)
	}
	if b.MaxTickMs != 200 {
		t.Errorf("expected MaxTickMs 200, got %d", b.MaxTickMs)
	}
}

func TestBoundsValidate(t *testing.T) {
	cases := []struct {
		name    string
		b       Bounds
		wantErr bool
	}{
		{"valid", Bounds{CpuPercent: 50, MemoryMb: 512, MaxTickMs: 100}, false},
		{"cpu 0", Bounds{CpuPercent: 0, MemoryMb: 512, MaxTickMs: 100}, true},
		{"cpu 101", Bounds{CpuPercent: 101, MemoryMb: 512, MaxTickMs: 100}, true},
		{"memory 0", Bounds{CpuPercent: 50, MemoryMb: 0, MaxTickMs: 100}, true},
		{"tick 0", Bounds{CpuPercent: 50, MemoryMb: 512, MaxTickMs: 0}, true},
		{"cpu 100 ok", Bounds{CpuPercent: 100, MemoryMb: 1, MaxTickMs: 1}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.b.Validate()
			if tc.wantErr && err == nil {
				t.Errorf("expected validation error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestBoundsToEnvVars(t *testing.T) {
	b := Bounds{CpuPercent: 60, MemoryMb: 256, MaxTickMs: 150}
	ev := b.ToEnvVars()
	if ev["IGRIS_MAX_CPU_PERCENT"] != "60" {
		t.Errorf("expected IGRIS_MAX_CPU_PERCENT=60, got %s", ev["IGRIS_MAX_CPU_PERCENT"])
	}
	if ev["IGRIS_MAX_MEMORY_MB"] != "256" {
		t.Errorf("expected IGRIS_MAX_MEMORY_MB=256, got %s", ev["IGRIS_MAX_MEMORY_MB"])
	}
	if ev["IGRIS_MAX_TICK_MS"] != "150" {
		t.Errorf("expected IGRIS_MAX_TICK_MS=150, got %s", ev["IGRIS_MAX_TICK_MS"])
	}
}

// -- ViolationRecord tests -----------------------------------------------

func sampleViolationRecord() ViolationRecord {
	return ViolationRecord{
		ID:            "0191f4a0-0000-7000-8000-000000000001",
		Timestamp:     time.Date(2026, 2, 18, 9, 0, 0, 0, time.UTC),
		ViolationKind: ViolationKindTime,
		Context:       map[string]interface{}{"task": "chat"},
		PreviousHash:  "",
		Hash:          "deadbeefdeadbeef",
		Signature:     "c2lnbmF0dXJl",
	}
}

func TestViolationRecordJSON(t *testing.T) {
	r := sampleViolationRecord()
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var r2 ViolationRecord
	if err := json.Unmarshal(data, &r2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if r2.ViolationKind != ViolationKindTime {
		t.Errorf("expected ViolationKindTime, got %v", r2.ViolationKind)
	}
	if r2.Hash != "deadbeefdeadbeef" {
		t.Errorf("expected hash deadbeefdeadbeef, got %s", r2.Hash)
	}
	if r2.Signature != "c2lnbmF0dXJl" {
		t.Errorf("expected signature c2lnbmF0dXJl, got %s", r2.Signature)
	}
}

// -- Runtime bounds integration ------------------------------------------

func TestRuntimeWithBoundsHeaderSent(t *testing.T) {
	var receivedHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Igris-Bounds")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "healthy"})
	}))
	defer server.Close()

	bounds := Bounds{CpuPercent: 40, MemoryMb: 256, MaxTickMs: 100}
	rt := NewRuntime(server.URL, WithBounds(bounds))
	_, _ = rt.Health(context.Background())

	if receivedHeader == "" {
		t.Fatal("expected X-Igris-Bounds header to be set, got empty")
	}

	var parsed map[string]int
	if err := json.Unmarshal([]byte(receivedHeader), &parsed); err != nil {
		t.Fatalf("X-Igris-Bounds is not valid JSON: %v", err)
	}
	if parsed["cpu_percent"] != 40 {
		t.Errorf("expected cpu_percent 40, got %d", parsed["cpu_percent"])
	}
	if parsed["memory_mb"] != 256 {
		t.Errorf("expected memory_mb 256, got %d", parsed["memory_mb"])
	}
	if parsed["max_tick_ms"] != 100 {
		t.Errorf("expected max_tick_ms 100, got %d", parsed["max_tick_ms"])
	}
}

func TestRuntimeNoBoundsHeaderWithoutConfig(t *testing.T) {
	var receivedHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Igris-Bounds")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "healthy"})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL) // no bounds
	_, _ = rt.Health(context.Background())

	if receivedHeader != "" {
		t.Errorf("expected no X-Igris-Bounds header, got %q", receivedHeader)
	}
}

func TestGetLastViolationReturnsLast(t *testing.T) {
	rec1 := sampleViolationRecord()
	rec2 := sampleViolationRecord()
	rec2.ID = "id-2"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/runtime/violations" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"violations": []ViolationRecord{rec1, rec2},
			})
		}
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	got, err := rt.GetLastViolation(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected a violation, got nil")
	}
	if got.ID != "id-2" {
		t.Errorf("expected id id-2, got %s", got.ID)
	}
}

func TestGetLastViolationNilWhenEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"violations": []ViolationRecord{}})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	got, err := rt.GetLastViolation(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestStreamViolationsReceivesNewRecords(t *testing.T) {
	calls := 0
	rec := sampleViolationRecord()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var violations []ViolationRecord
		if calls >= 2 {
			violations = []ViolationRecord{rec}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"violations": violations})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch := rt.StreamViolations(ctx)
	select {
	case v := <-ch:
		if v.ID != rec.ID {
			t.Errorf("expected violation ID %s, got %s", rec.ID, v.ID)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for violation")
	}
}
