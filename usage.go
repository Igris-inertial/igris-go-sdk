package igris

import (
	"context"
	"net/http"
)

// UsageManager accesses usage data.
type UsageManager struct {
	client *Client
}

// Current returns current usage data.
func (m *UsageManager) Current(ctx context.Context) (*Usage, error) {
	var resp Usage
	err := m.client.doRequest(ctx, http.MethodGet, "/v1/usage", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// History returns usage history.
func (m *UsageManager) History(ctx context.Context, params map[string]string) (*UsageHistory, error) {
	var resp UsageHistory
	err := m.client.doRequestWithParams(ctx, http.MethodGet, "/v1/usage/history", params, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AuditManager accesses audit logs.
type AuditManager struct {
	client *Client
}

// List returns audit log entries.
func (m *AuditManager) List(ctx context.Context, params map[string]string) ([]AuditEntry, error) {
	var resp struct {
		Entries []AuditEntry `json:"entries"`
	}
	err := m.client.doRequestWithParams(ctx, http.MethodGet, "/v1/audit", params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Entries, nil
}
