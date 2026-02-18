package igris

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RuntimeConfig holds configuration for connecting to the local Igris Runtime.
type RuntimeConfig struct {
	LocalURL     string
	CloudURL     string
	AutoFallback bool
	SocketPath   string
	Timeout      time.Duration
	LocalModel   string
}

// RuntimeOption configures a Runtime instance.
type RuntimeOption func(*Runtime)

// WithCloudURL sets the cloud fallback URL.
func WithCloudURL(url string) RuntimeOption {
	return func(r *Runtime) {
		r.cloudURL = strings.TrimRight(url, "/")
	}
}

// WithAutoFallback enables or disables automatic cloud fallback.
func WithAutoFallback(enabled bool) RuntimeOption {
	return func(r *Runtime) {
		r.autoFallback = enabled
	}
}

// WithRuntimeTimeout sets the HTTP timeout for runtime requests.
func WithRuntimeTimeout(d time.Duration) RuntimeOption {
	return func(r *Runtime) {
		r.httpClient.Timeout = d
	}
}

// WithLocalModel sets the default local model.
func WithLocalModel(model string) RuntimeOption {
	return func(r *Runtime) {
		r.localModel = model
	}
}

// WithBounds sets containment bounds that are propagated to the runtime via
// the X-Igris-Bounds request header on every local and cloud request.
func WithBounds(b Bounds) RuntimeOption {
	return func(r *Runtime) {
		r.bounds = &b
	}
}

// Runtime is a client for the local Igris Runtime binary.
type Runtime struct {
	localURL     string
	cloudURL     string
	autoFallback bool
	localModel   string
	httpClient   *http.Client
	bounds       *Bounds

	// Violation observability
	mu                   sync.Mutex
	lastSeenViolationID  string
	violationCallbacks   []func(ViolationRecord)
	stopPoller           chan struct{}
}

// NewRuntime creates a new Runtime client.
func NewRuntime(localURL string, opts ...RuntimeOption) *Runtime {
	r := &Runtime{
		localURL:     strings.TrimRight(localURL, "/"),
		autoFallback: true,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Runtime) doLocalRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	return r.doHTTPRequest(ctx, r.localURL, method, path, body, result)
}

func (r *Runtime) doCloudRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	if r.cloudURL == "" {
		return &NetworkError{Err: fmt.Errorf("no cloud URL configured for fallback")}
	}
	return r.doHTTPRequest(ctx, r.cloudURL, method, path, body, result)
}

func (r *Runtime) doHTTPRequest(ctx context.Context, baseURL, method, path string, body interface{}, result interface{}) error {
	url := baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return &NetworkError{Err: fmt.Errorf("create request: %w", err)}
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if r.bounds != nil {
		if b, err2 := json.Marshal(map[string]int{
			"cpu_percent": r.bounds.CpuPercent,
			"memory_mb":   r.bounds.MemoryMb,
			"max_tick_ms": r.bounds.MaxTickMs,
		}); err2 == nil {
			req.Header.Set("X-Igris-Bounds", string(b))
		}
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return &NetworkError{Err: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &NetworkError{Err: err}
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return &AuthenticationError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}
	if resp.StatusCode == 429 {
		return &RateLimitError{Message: string(respBody)}
	}
	if resp.StatusCode == 400 || resp.StatusCode == 422 {
		return &ValidationError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}
	if resp.StatusCode >= 400 {
		return &APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	if result != nil && resp.StatusCode != 204 && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}

func (r *Runtime) doRequestWithFallback(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	err := r.doLocalRequest(ctx, method, path, body, result)
	if err != nil {
		if _, ok := err.(*NetworkError); ok && r.autoFallback && r.cloudURL != "" {
			return r.doCloudRequest(ctx, method, path, body, result)
		}
		return err
	}
	return nil
}

// Chat sends a chat completion request to the runtime with automatic cloud fallback.
func (r *Runtime) Chat(ctx context.Context, req *InferRequest) (*InferResponse, error) {
	var resp InferResponse
	if err := r.doRequestWithFallback(ctx, http.MethodPost, "/v1/chat/completions", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ChatLocal sends a chat completion request only to the local runtime (no fallback).
func (r *Runtime) ChatLocal(ctx context.Context, req *InferRequest) (*InferResponse, error) {
	var resp InferResponse
	if err := r.doLocalRequest(ctx, http.MethodPost, "/v1/chat/completions", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// LoadModel loads a GGUF model into the runtime.
func (r *Runtime) LoadModel(ctx context.Context, modelPath string, modelID string) (map[string]interface{}, error) {
	payload := map[string]string{"model_path": modelPath}
	if modelID != "" {
		payload["model_id"] = modelID
	}
	var result map[string]interface{}
	if err := r.doLocalRequest(ctx, http.MethodPost, "/v1/admin/models/load", payload, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SwapModel hot-swaps to a different loaded model.
func (r *Runtime) SwapModel(ctx context.Context, modelID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := r.doLocalRequest(ctx, http.MethodPost, "/v1/admin/models/swap", map[string]string{"model_id": modelID}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListModels lists models available in the runtime.
func (r *Runtime) ListModels(ctx context.Context) ([]map[string]interface{}, error) {
	var result struct {
		Models []map[string]interface{} `json:"models"`
	}
	if err := r.doLocalRequest(ctx, http.MethodGet, "/v1/admin/models", nil, &result); err != nil {
		return nil, err
	}
	return result.Models, nil
}

// Health checks runtime health status.
func (r *Runtime) Health(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := r.doLocalRequest(ctx, http.MethodGet, "/v1/health", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// -- Containment observability --------------------------------------------

func (r *Runtime) fetchViolations(ctx context.Context) ([]ViolationRecord, error) {
	var result struct {
		Violations []ViolationRecord `json:"violations"`
	}
	if err := r.doLocalRequest(ctx, http.MethodGet, "/v1/runtime/violations", nil, &result); err != nil {
		return nil, err
	}
	return result.Violations, nil
}

// GetLastViolation returns the most recent violation record from the runtime,
// or nil if none exist or the endpoint is unavailable.
func (r *Runtime) GetLastViolation(ctx context.Context) (*ViolationRecord, error) {
	records, err := r.fetchViolations(ctx)
	if err != nil {
		return nil, nil // endpoint not yet wired; treat as no violations
	}
	if len(records) == 0 {
		return nil, nil
	}
	last := records[len(records)-1]
	return &last, nil
}

// OnViolation registers a callback that fires for each new violation record.
// A background goroutine polls GET /v1/runtime/violations every second.
// Call StopViolationPoller to stop it.
func (r *Runtime) OnViolation(callback func(ViolationRecord)) {
	r.mu.Lock()
	r.violationCallbacks = append(r.violationCallbacks, callback)
	alreadyRunning := r.stopPoller != nil
	r.mu.Unlock()

	if alreadyRunning {
		return
	}

	stop := make(chan struct{})
	r.mu.Lock()
	r.stopPoller = stop
	r.mu.Unlock()

	go r.pollLoop(stop)
}

func (r *Runtime) pollLoop(stop chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			records, err := r.fetchViolations(context.Background())
			if err != nil || len(records) == 0 {
				continue
			}
			last := records[len(records)-1]

			r.mu.Lock()
			seen := r.lastSeenViolationID == last.ID
			if !seen {
				r.lastSeenViolationID = last.ID
			}
			cbs := r.violationCallbacks
			r.mu.Unlock()

			if !seen {
				for _, cb := range cbs {
					cb(last)
				}
			}
		}
	}
}

// StopViolationPoller stops the background polling goroutine started by OnViolation.
func (r *Runtime) StopViolationPoller() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stopPoller != nil {
		close(r.stopPoller)
		r.stopPoller = nil
	}
}

// StreamViolations returns a channel that receives new violation records as they
// are detected (polling every second). Close ctx to stop streaming.
func (r *Runtime) StreamViolations(ctx context.Context) <-chan ViolationRecord {
	ch := make(chan ViolationRecord, 8)
	go func() {
		defer close(ch)
		seen := make(map[string]struct{})
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				records, err := r.fetchViolations(ctx)
				if err != nil {
					continue
				}
				for _, rec := range records {
					if _, ok := seen[rec.ID]; !ok {
						seen[rec.ID] = struct{}{}
						select {
						case ch <- rec:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()
	return ch
}
