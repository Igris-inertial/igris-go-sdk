package igris

import (
	"context"
	"fmt"
	"net/http"
)

// ProviderManager manages inference providers.
type ProviderManager struct {
	client *Client
}

// Register registers a new provider.
func (m *ProviderManager) Register(ctx context.Context, config *ProviderConfig) (*Provider, error) {
	var resp Provider
	err := m.client.doRequest(ctx, http.MethodPost, "/v1/providers/register", config, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// List returns all registered providers.
func (m *ProviderManager) List(ctx context.Context) ([]Provider, error) {
	var resp struct {
		Providers []Provider `json:"providers"`
	}
	err := m.client.doRequest(ctx, http.MethodGet, "/v1/providers", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Providers, nil
}

// Test tests a provider configuration.
func (m *ProviderManager) Test(ctx context.Context, config *ProviderConfig) (*TestResult, error) {
	var resp TestResult
	err := m.client.doRequest(ctx, http.MethodPost, "/v1/providers/test", config, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update updates a provider.
func (m *ProviderManager) Update(ctx context.Context, id string, config map[string]interface{}) (*Provider, error) {
	var resp Provider
	err := m.client.doRequest(ctx, http.MethodPut, fmt.Sprintf("/v1/providers/%s", id), config, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Delete removes a provider.
func (m *ProviderManager) Delete(ctx context.Context, id string) error {
	return m.client.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/v1/providers/%s", id), nil, nil)
}

// Health checks a provider's health.
func (m *ProviderManager) Health(ctx context.Context, id string) (*HealthStatus, error) {
	var resp HealthStatus
	err := m.client.doRequest(ctx, http.MethodGet, fmt.Sprintf("/v1/providers/%s/health", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
