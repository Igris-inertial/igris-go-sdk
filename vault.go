package igris

import (
	"context"
	"fmt"
	"net/http"
)

// VaultManager manages BYOK vault keys.
type VaultManager struct {
	client *Client
}

// Store stores a key in the vault.
func (m *VaultManager) Store(ctx context.Context, req *VaultStoreRequest) (*VaultKey, error) {
	var resp VaultKey
	err := m.client.doRequest(ctx, http.MethodPost, "/v1/vault/keys", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// List returns all vault keys.
func (m *VaultManager) List(ctx context.Context) ([]VaultKey, error) {
	var resp struct {
		Keys []VaultKey `json:"keys"`
	}
	err := m.client.doRequest(ctx, http.MethodGet, "/v1/vault/keys", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Keys, nil
}

// Rotate rotates a provider's key.
func (m *VaultManager) Rotate(ctx context.Context, provider string) (*VaultKey, error) {
	var resp VaultKey
	err := m.client.doRequest(ctx, http.MethodPost, fmt.Sprintf("/v1/vault/keys/%s/rotate", provider), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Delete removes a provider's key.
func (m *VaultManager) Delete(ctx context.Context, provider string) error {
	return m.client.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/v1/vault/keys/%s", provider), nil, nil)
}
