package igris

// Message represents a chat message.
type Message struct {
	Role         string                   `json:"role"`
	Content      string                   `json:"content"`
	ContentParts []map[string]interface{} `json:"content_parts,omitempty"`
}

// PolicyOverride configures routing policy.
type PolicyOverride struct {
	Provider        string `json:"provider,omitempty"`
	OptimizeFor     string `json:"optimize_for,omitempty"`
	FallbackEnabled *bool  `json:"fallback_enabled,omitempty"`
	TimeoutMs       int    `json:"timeout_ms,omitempty"`
}

// InferRequest is the main inference request.
type InferRequest struct {
	Model       string            `json:"model"`
	Messages    []Message         `json:"messages"`
	Stream      *bool             `json:"stream,omitempty"`
	MaxTokens   *int              `json:"max_tokens,omitempty"`
	Temperature *float64          `json:"temperature,omitempty"`
	TopP        *float64          `json:"top_p,omitempty"`
	Stop        []string          `json:"stop,omitempty"`
	Policy      *PolicyOverride   `json:"policy,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// UsageStats contains token usage information.
type UsageStats struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ResponseMetadata contains routing metadata.
type ResponseMetadata struct {
	Provider  string  `json:"provider,omitempty"`
	LatencyMs float64 `json:"latency_ms,omitempty"`
	CostUSD   float64 `json:"cost_usd,omitempty"`
	CacheHit  bool    `json:"cache_hit,omitempty"`
}

// Choice represents a completion choice.
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason,omitempty"`
}

// ExecutionReceipt is the resource-accounting receipt produced by a Runtime
// instance after each execution. It is signed with the Runtime's Ed25519 key
// and hash-chained for tamper evidence. Present only when the Runtime emits it.
type ExecutionReceipt struct {
	ExecutionID      string `json:"execution_id"`
	AgentID          string `json:"agent_id,omitempty"`
	TransactionID    string `json:"transaction_id,omitempty"`
	TransactionHash  string `json:"transaction_hash,omitempty"`
	CpuTimeMs        int64  `json:"cpu_time_ms"`
	WallTimeMs       int64  `json:"wall_time_ms"`
	MemoryPeakMb     int64  `json:"memory_peak_mb"`
	FsBytesWritten   int64  `json:"fs_bytes_written"`
	ToolCalls        int    `json:"tool_calls"`
	ViolationOccurred bool  `json:"violation_occurred"`
	TimestampUTC     string `json:"timestamp_utc,omitempty"`
	PreviousHash     string `json:"previous_hash,omitempty"`
	Hash             string `json:"hash"`
	Signature        string `json:"signature"`
}

// InferResponse is the inference response.
type InferResponse struct {
	ID               string            `json:"id"`
	Object           string            `json:"object"`
	Created          int64             `json:"created"`
	Model            string            `json:"model"`
	Choices          []Choice          `json:"choices"`
	Usage            *UsageStats       `json:"usage,omitempty"`
	Metadata         *ResponseMetadata `json:"metadata,omitempty"`
	ExecutionReceipt *ExecutionReceipt `json:"execution_receipt,omitempty"`
}

// ModelsResponse lists available models.
type ModelsResponse struct {
	Models []map[string]interface{} `json:"models"`
}

// HealthResponse is the health check response.
type HealthResponse struct {
	Status  string  `json:"status"`
	Version string  `json:"version,omitempty"`
	Uptime  float64 `json:"uptime,omitempty"`
}

// ProviderConfig configures a provider.
type ProviderConfig struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	APIKey   string   `json:"api_key,omitempty"`
	BaseURL  string   `json:"base_url,omitempty"`
	Models   []string `json:"models,omitempty"`
	Priority int      `json:"priority,omitempty"`
	Weight   float64  `json:"weight,omitempty"`
	Enabled  *bool    `json:"enabled,omitempty"`
}

// Provider represents a registered provider.
type Provider struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Models  []string `json:"models,omitempty"`
	Enabled bool     `json:"enabled"`
	Status  string   `json:"status,omitempty"`
}

// TestResult is the result of testing a provider.
type TestResult struct {
	Success   bool    `json:"success"`
	LatencyMs float64 `json:"latency_ms,omitempty"`
	Error     string  `json:"error,omitempty"`
}

// HealthStatus is a provider's health status.
type HealthStatus struct {
	ProviderID string  `json:"provider_id"`
	Status     string  `json:"status"`
	LatencyMs  float64 `json:"latency_ms,omitempty"`
	LastCheck  string  `json:"last_check,omitempty"`
}

// VaultKey represents a stored vault key.
type VaultKey struct {
	Provider  string `json:"provider"`
	KeyID     string `json:"key_id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	RotatedAt string `json:"rotated_at,omitempty"`
}

// VaultStoreRequest stores a key in the vault.
type VaultStoreRequest struct {
	Provider string                 `json:"provider"`
	APIKey   string                 `json:"api_key"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// Usage contains current usage data.
type Usage struct {
	TotalRequests int     `json:"total_requests"`
	TotalTokens   int     `json:"total_tokens"`
	TotalCostUSD  float64 `json:"total_cost_usd"`
	Period        string  `json:"period,omitempty"`
}

// UsageHistory contains historical usage data.
type UsageHistory struct {
	Entries []map[string]interface{} `json:"entries"`
	Period  string                   `json:"period,omitempty"`
}

// AuditEntry represents an audit log entry.
type AuditEntry struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	Timestamp string                 `json:"timestamp"`
	User      string                 `json:"user,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// FleetAgent represents a registered fleet agent.
type FleetAgent struct {
	ID           string   `json:"id,omitempty"`
	FleetID      string   `json:"fleet_id,omitempty"`
	Name         string   `json:"name"`
	Status       string   `json:"status,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	RegisteredAt string   `json:"registered_at,omitempty"`
}

// FleetHealth represents fleet health status.
type FleetHealth struct {
	TotalAgents int                      `json:"total_agents"`
	Healthy     int                      `json:"healthy"`
	Unhealthy   int                      `json:"unhealthy"`
	Agents      []map[string]interface{} `json:"agents,omitempty"`
}
