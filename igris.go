// Package igris provides a client for the Igris Inertial AI inference gateway.
package igris

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client for the Igris Inertial AI inference gateway.
type Client struct {
	baseURL    string
	apiKey     string
	tenantID   string
	httpClient *http.Client
	token      string

	Providers *ProviderManager
	Vault     *VaultManager
	Fleet     *FleetManager
	Usage     *UsageManager
	Audit     *AuditManager
}

// Option configures the Client.
type Option func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithTenantID sets the tenant ID header.
func WithTenantID(id string) Option {
	return func(c *Client) {
		c.tenantID = id
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// NewClient creates a new Igris Inertial client.
func NewClient(baseURL, apiKey string, opts ...Option) *Client {
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	c.Providers = &ProviderManager{client: c}
	c.Vault = &VaultManager{client: c}
	c.Fleet = &FleetManager{client: c}
	c.Usage = &UsageManager{client: c}
	c.Audit = &AuditManager{client: c}
	return c
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	u := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	auth := c.token
	if auth == "" {
		auth = c.apiKey
	}
	if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}
	if c.tenantID != "" {
		req.Header.Set("X-Tenant-ID", c.tenantID)
	}

	resp, err := c.httpClient.Do(req)
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

func (c *Client) doRequestWithParams(ctx context.Context, method, path string, params map[string]string, result interface{}) error {
	if len(params) > 0 {
		v := url.Values{}
		for k, val := range params {
			v.Set(k, val)
		}
		path = path + "?" + v.Encode()
	}
	return c.doRequest(ctx, method, path, nil, result)
}

// Login authenticates and stores the JWT token.
func (c *Client) Login(ctx context.Context, apiKey string) (map[string]interface{}, error) {
	if apiKey == "" {
		apiKey = c.apiKey
	}
	var result map[string]interface{}
	err := c.doRequest(ctx, http.MethodPost, "/v1/auth/login", map[string]string{"api_key": apiKey}, &result)
	if err != nil {
		return nil, err
	}
	if token, ok := result["token"].(string); ok {
		c.token = token
	}
	return result, nil
}

// RefreshToken refreshes the JWT.
func (c *Client) RefreshToken(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.doRequest(ctx, http.MethodPost, "/v1/auth/refresh", nil, &result)
	if err != nil {
		return nil, err
	}
	if token, ok := result["token"].(string); ok {
		c.token = token
	}
	return result, nil
}

// Logout invalidates the current token.
func (c *Client) Logout(ctx context.Context) error {
	err := c.doRequest(ctx, http.MethodPost, "/v1/auth/logout", nil, nil)
	c.token = ""
	return err
}

// Infer sends an inference request.
func (c *Client) Infer(ctx context.Context, req *InferRequest) (*InferResponse, error) {
	var resp InferResponse
	err := c.doRequest(ctx, http.MethodPost, "/v1/infer", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ChatCompletion sends a chat completion request (OpenAI-compatible).
func (c *Client) ChatCompletion(ctx context.Context, req *InferRequest) (*InferResponse, error) {
	var resp InferResponse
	err := c.doRequest(ctx, http.MethodPost, "/v1/chat/completions", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListModels returns available models.
func (c *Client) ListModels(ctx context.Context) (*ModelsResponse, error) {
	var resp ModelsResponse
	err := c.doRequest(ctx, http.MethodGet, "/v1/models", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Health checks gateway health.
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	var resp HealthResponse
	err := c.doRequest(ctx, http.MethodGet, "/v1/health", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ProviderStats returns provider statistics.
func (c *Client) ProviderStats(ctx context.Context) (map[string]interface{}, error) {
	var resp map[string]interface{}
	err := c.doRequest(ctx, http.MethodGet, "/v1/providers/stats", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UploadKey stores a provider API key in the vault.
func (c *Client) UploadKey(ctx context.Context, provider, apiKey string) (*VaultKey, error) {
	return c.Vault.Store(ctx, &VaultStoreRequest{Provider: provider, APIKey: apiKey})
}

// RotateKey rotates a provider API key.
func (c *Client) RotateKey(ctx context.Context, provider string) (*VaultKey, error) {
	return c.Vault.Rotate(ctx, provider)
}

// ListKeys lists all stored vault keys.
func (c *Client) ListKeys(ctx context.Context) ([]VaultKey, error) {
	return c.Vault.List(ctx)
}
