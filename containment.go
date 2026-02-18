package igris

import (
	"fmt"
	"strconv"
	"time"
)

// Bounds specifies containment limits forwarded to the Igris Runtime.
//
// These mirror the Bounds struct inside igris-safety and control the CPU
// quota and per-tick time limit enforced on each inference worker process.
//
// Deterministic failure semantics:
//   - Time violation: worker exceeded MaxTickMs; supervisor sends SIGKILL,
//     writes a signed ViolationRecord to the audit log, and respawns the worker.
//   - CPU violation: worker exceeded the cgroup CPU quota; treated identically.
//   - Cloud outage: Runtime.Chat falls back to local when AutoFallback is true.
//   - Worker SIGKILL: supervisor reaps the process, destroys the cgroup, records
//     the violation, and spawns a new worker automatically.
type Bounds struct {
	// CpuPercent is the maximum CPU usage as a percentage (1–100).
	CpuPercent int
	// MemoryMb is the maximum RSS memory in megabytes (> 0).
	MemoryMb int
	// MaxTickMs is the hard deadline in milliseconds for each worker execution.
	MaxTickMs int
}

// DefaultBounds returns a sensible default Bounds value.
func DefaultBounds() Bounds {
	return Bounds{CpuPercent: 80, MemoryMb: 512, MaxTickMs: 200}
}

// Validate returns an error if any Bounds field is out of the valid range.
func (b Bounds) Validate() error {
	if b.CpuPercent < 1 || b.CpuPercent > 100 {
		return fmt.Errorf("igris: CpuPercent must be 1–100, got %d", b.CpuPercent)
	}
	if b.MemoryMb <= 0 {
		return fmt.Errorf("igris: MemoryMb must be > 0, got %d", b.MemoryMb)
	}
	if b.MaxTickMs <= 0 {
		return fmt.Errorf("igris: MaxTickMs must be > 0, got %d", b.MaxTickMs)
	}
	return nil
}

// ToEnvVars returns a map of IGRIS_MAX_* env-var names to string values for
// propagating bounds to a subprocess that launches the runtime.
func (b Bounds) ToEnvVars() map[string]string {
	return map[string]string{
		"IGRIS_MAX_CPU_PERCENT": strconv.Itoa(b.CpuPercent),
		"IGRIS_MAX_MEMORY_MB":   strconv.Itoa(b.MemoryMb),
		"IGRIS_MAX_TICK_MS":     strconv.Itoa(b.MaxTickMs),
	}
}

// ViolationKind identifies the type of containment violation.
type ViolationKind string

const (
	ViolationKindTime ViolationKind = "Time"
	ViolationKindCpu  ViolationKind = "Cpu"
)

// ViolationRecord is a parsed containment violation record from the Igris Runtime.
// The runtime writes one JSON line per violation; this struct mirrors that structure.
type ViolationRecord struct {
	ID            string            `json:"id"`
	Timestamp     time.Time         `json:"timestamp"`
	ViolationKind ViolationKind     `json:"violation_kind"`
	Context       map[string]interface{} `json:"context"`
	PreviousHash  string            `json:"previous_hash"`
	Hash          string            `json:"hash"`
	// Signature is a base64-encoded Ed25519 signature over Hash.
	Signature     string            `json:"signature"`
}
