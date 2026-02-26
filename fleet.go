package igris

import (
	"context"
	"fmt"
	"net/http"
)

// FleetManager manages fleet agents.
type FleetManager struct {
	client *Client
}

// Register registers a new fleet agent.
func (m *FleetManager) Register(ctx context.Context, config map[string]interface{}) (*FleetAgent, error) {
	var resp FleetAgent
	err := m.client.doRequest(ctx, http.MethodPost, "/api/fleet/register", config, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Telemetry uploads telemetry data.
func (m *FleetManager) Telemetry(ctx context.Context, fleetID string, data map[string]interface{}) error {
	return m.client.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/fleet/%s/telemetry", fleetID), data, nil)
}

// Agents lists all fleet agents.
func (m *FleetManager) Agents(ctx context.Context) ([]FleetAgent, error) {
	var resp struct {
		Agents []FleetAgent `json:"agents"`
	}
	err := m.client.doRequest(ctx, http.MethodGet, "/api/fleet/agents", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Agents, nil
}

// Health returns fleet health status.
func (m *FleetManager) Health(ctx context.Context) (*FleetHealth, error) {
	var resp FleetHealth
	err := m.client.doRequest(ctx, http.MethodGet, "/api/fleet/health", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
