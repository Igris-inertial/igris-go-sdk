package igris

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewRuntime(t *testing.T) {
	r := NewRuntime("http://localhost:9090")
	if r.localURL != "http://localhost:9090" {
		t.Errorf("expected local URL http://localhost:9090, got %s", r.localURL)
	}
	if r.autoFallback != true {
		t.Error("expected autoFallback to default to true")
	}
	if r.cloudURL != "" {
		t.Errorf("expected empty cloud URL, got %s", r.cloudURL)
	}
	if r.localModel != "" {
		t.Errorf("expected empty local model, got %s", r.localModel)
	}
}

func TestNewRuntimeTrimsTrailingSlash(t *testing.T) {
	r := NewRuntime("http://localhost:9090/")
	if r.localURL != "http://localhost:9090" {
		t.Errorf("expected trailing slash trimmed, got %s", r.localURL)
	}
}

func TestNewRuntimeWithOptions(t *testing.T) {
	r := NewRuntime("http://localhost:9090",
		WithCloudURL("https://cloud.igris.dev/"),
		WithAutoFallback(false),
		WithRuntimeTimeout(5*time.Second),
		WithLocalModel("llama-7b"),
	)
	if r.cloudURL != "https://cloud.igris.dev" {
		t.Errorf("expected cloud URL https://cloud.igris.dev, got %s", r.cloudURL)
	}
	if r.autoFallback != false {
		t.Error("expected autoFallback to be false")
	}
	if r.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", r.httpClient.Timeout)
	}
	if r.localModel != "llama-7b" {
		t.Errorf("expected local model llama-7b, got %s", r.localModel)
	}
}

func TestRuntimeChatLocalSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("expected path /v1/chat/completions, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		resp := InferResponse{
			ID:      "local-123",
			Object:  "chat.completion",
			Created: 1700000000,
			Model:   "llama-7b",
			Choices: []Choice{
				{Index: 0, Message: Message{Role: "assistant", Content: "Hello from local!"}, FinishReason: "stop"},
			},
			Usage: &UsageStats{PromptTokens: 8, CompletionTokens: 4, TotalTokens: 12},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	resp, err := rt.ChatLocal(context.Background(), &InferRequest{
		Model:    "llama-7b",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "local-123" {
		t.Errorf("expected id local-123, got %s", resp.ID)
	}
	if resp.Choices[0].Message.Content != "Hello from local!" {
		t.Errorf("expected content Hello from local!, got %s", resp.Choices[0].Message.Content)
	}
	if resp.Usage.TotalTokens != 12 {
		t.Errorf("expected 12 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestRuntimeChatFallbackOnConnectionError(t *testing.T) {
	cloudServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := InferResponse{
			ID:      "cloud-456",
			Object:  "chat.completion",
			Created: 1700000000,
			Model:   "gpt-4",
			Choices: []Choice{
				{Index: 0, Message: Message{Role: "assistant", Content: "Hello from cloud!"}, FinishReason: "stop"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer cloudServer.Close()

	// Use an invalid local URL to force a connection error.
	rt := NewRuntime("http://127.0.0.1:1", WithCloudURL(cloudServer.URL), WithRuntimeTimeout(1*time.Second))
	resp, err := rt.Chat(context.Background(), &InferRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	if err != nil {
		t.Fatalf("expected fallback to succeed, got error: %v", err)
	}
	if resp.ID != "cloud-456" {
		t.Errorf("expected cloud response id cloud-456, got %s", resp.ID)
	}
	if resp.Choices[0].Message.Content != "Hello from cloud!" {
		t.Errorf("expected cloud content, got %s", resp.Choices[0].Message.Content)
	}
}

func TestRuntimeChatFallbackDisabled(t *testing.T) {
	cloudServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := InferResponse{ID: "cloud-789"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer cloudServer.Close()

	// Disable fallback; should get a NetworkError instead of falling back to cloud.
	rt := NewRuntime("http://127.0.0.1:1",
		WithCloudURL(cloudServer.URL),
		WithAutoFallback(false),
		WithRuntimeTimeout(1*time.Second),
	)
	_, err := rt.Chat(context.Background(), &InferRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	if err == nil {
		t.Fatal("expected error when fallback is disabled")
	}
	if _, ok := err.(*NetworkError); !ok {
		t.Errorf("expected NetworkError, got %T: %v", err, err)
	}
}

func TestRuntimeLoadModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/admin/models/load" {
			t.Errorf("expected path /v1/admin/models/load, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["model_path"] != "/models/llama-7b.gguf" {
			t.Errorf("expected model_path /models/llama-7b.gguf, got %s", body["model_path"])
		}
		if body["model_id"] != "llama-7b" {
			t.Errorf("expected model_id llama-7b, got %s", body["model_id"])
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "loaded",
			"model_id": "llama-7b",
		})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	result, err := rt.LoadModel(context.Background(), "/models/llama-7b.gguf", "llama-7b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["status"] != "loaded" {
		t.Errorf("expected status loaded, got %v", result["status"])
	}
}

func TestRuntimeSwapModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/admin/models/swap" {
			t.Errorf("expected path /v1/admin/models/swap, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["model_id"] != "mistral-7b" {
			t.Errorf("expected model_id mistral-7b, got %s", body["model_id"])
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "swapped",
			"model_id": "mistral-7b",
		})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	result, err := rt.SwapModel(context.Background(), "mistral-7b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["status"] != "swapped" {
		t.Errorf("expected status swapped, got %v", result["status"])
	}
}

func TestRuntimeListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/admin/models" {
			t.Errorf("expected path /v1/admin/models, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"models": []map[string]interface{}{
				{"id": "llama-7b", "status": "loaded"},
				{"id": "mistral-7b", "status": "ready"},
			},
		})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	models, err := rt.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
	if models[0]["id"] != "llama-7b" {
		t.Errorf("expected first model llama-7b, got %v", models[0]["id"])
	}
	if models[1]["id"] != "mistral-7b" {
		t.Errorf("expected second model mistral-7b, got %v", models[1]["id"])
	}
}

func TestRuntimeHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/health" {
			t.Errorf("expected path /v1/health, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"version": "0.5.0",
			"uptime":  3600.0,
		})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	result, err := rt.Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["status"] != "healthy" {
		t.Errorf("expected status healthy, got %v", result["status"])
	}
	if result["version"] != "0.5.0" {
		t.Errorf("expected version 0.5.0, got %v", result["version"])
	}
}

func TestRuntimeValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"error":"invalid model format"}`))
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	_, err := rt.Chat(context.Background(), &InferRequest{
		Model:    "",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	if ve.StatusCode != 422 {
		t.Errorf("expected status code 422, got %d", ve.StatusCode)
	}
}
