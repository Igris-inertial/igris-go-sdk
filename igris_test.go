package igris

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:8080", "test-key")
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("expected base URL http://localhost:8080, got %s", c.baseURL)
	}
	if c.apiKey != "test-key" {
		t.Errorf("expected api key test-key, got %s", c.apiKey)
	}
	if c.Providers == nil || c.Vault == nil || c.Fleet == nil || c.Usage == nil || c.Audit == nil {
		t.Error("expected all managers to be initialized")
	}
}

func TestClientWithOptions(t *testing.T) {
	c := NewClient("http://localhost:8080", "key", WithTenantID("t1"))
	if c.tenantID != "t1" {
		t.Errorf("expected tenant ID t1, got %s", c.tenantID)
	}
}

func TestInfer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/infer" {
			t.Errorf("expected path /v1/infer, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer test-key auth header")
		}

		resp := InferResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1700000000,
			Model:   "gpt-4",
			Choices: []Choice{
				{Index: 0, Message: Message{Role: "assistant", Content: "Hello!"}, FinishReason: "stop"},
			},
			Usage: &UsageStats{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	resp, err := c.Infer(context.Background(), &InferRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "chatcmpl-123" {
		t.Errorf("expected id chatcmpl-123, got %s", resp.ID)
	}
	if resp.Choices[0].Message.Content != "Hello!" {
		t.Errorf("expected content Hello!, got %s", resp.Choices[0].Message.Content)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("expected 15 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(HealthResponse{Status: "healthy", Version: "1.0.0"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key")
	health, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.Status != "healthy" {
		t.Errorf("expected healthy, got %s", health.Status)
	}
}

func TestListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"models": []map[string]interface{}{{"id": "gpt-4"}, {"id": "claude-3-opus"}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key")
	models, err := c.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models.Models) != 2 {
		t.Errorf("expected 2 models, got %d", len(models.Models))
	}
}

func TestAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "bad-key")
	_, err := c.Health(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Errorf("expected AuthenticationError, got %T", err)
	}
}

func TestRateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "key")
	_, err := c.Infer(context.Background(), &InferRequest{Model: "gpt-4", Messages: []Message{{Role: "user", Content: "hi"}}})
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*RateLimitError); !ok {
		t.Errorf("expected RateLimitError, got %T", err)
	}
}

func TestProvidersList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"providers": []map[string]interface{}{
				{"id": "p1", "name": "OpenAI", "type": "openai", "enabled": true},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key")
	providers, err := c.Providers.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(providers))
	}
	if providers[0].Name != "OpenAI" {
		t.Errorf("expected OpenAI, got %s", providers[0].Name)
	}
}

func TestVaultStore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(VaultKey{Provider: "openai", KeyID: "k1"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key")
	key, err := c.Vault.Store(context.Background(), &VaultStoreRequest{Provider: "openai", APIKey: "sk-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key.Provider != "openai" {
		t.Errorf("expected openai, got %s", key.Provider)
	}
}
