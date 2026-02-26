package igris

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewBehaviorTree(t *testing.T) {
	rt := NewRuntime("http://localhost:9090")
	tree := map[string]interface{}{
		"type": "sequence",
		"name": "root",
	}
	bt := NewBehaviorTree(tree, rt)
	if bt.runtime != rt {
		t.Error("expected runtime to be set")
	}
	if bt.tree["type"] != "sequence" {
		t.Errorf("expected tree type sequence, got %v", bt.tree["type"])
	}
	if bt.tree["name"] != "root" {
		t.Errorf("expected tree name root, got %v", bt.tree["name"])
	}
}

func TestBTreeValidateSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/btree/validate" {
			t.Errorf("expected path /v1/btree/validate, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["tree"] == nil {
			t.Error("expected tree field in request body")
		}

		json.NewEncoder(w).Encode(BTreeValidateResult{
			Valid:    true,
			RootType: "sequence",
			RootName: "main_sequence",
		})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	tree := NewSequenceNode("main_sequence",
		NewActionNode("greet", "say_hello", map[string]interface{}{"target": "world"}),
	)
	bt := NewBehaviorTree(tree, rt)

	result, err := bt.Validate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Error("expected valid to be true")
	}
	if result.RootType != "sequence" {
		t.Errorf("expected root_type sequence, got %s", result.RootType)
	}
	if result.RootName != "main_sequence" {
		t.Errorf("expected root_name main_sequence, got %s", result.RootName)
	}
}

func TestBTreeValidateInvalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(BTreeValidateResult{
			Valid: false,
			Error: "missing root node",
		})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	bt := NewBehaviorTree(map[string]interface{}{}, rt)

	result, err := bt.Validate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected valid to be false")
	}
	if result.Error != "missing root node" {
		t.Errorf("expected error 'missing root node', got %s", result.Error)
	}
}

func TestBTreeValidateServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"error":"malformed tree definition"}`))
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	bt := NewBehaviorTree(map[string]interface{}{"bad": true}, rt)

	_, err := bt.Validate(context.Background())
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

func TestBTreeRunSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/btree/run" {
			t.Errorf("expected path /v1/btree/run, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["tree"] == nil {
			t.Error("expected tree field in request body")
		}

		json.NewEncoder(w).Encode(BTreeRunResult{
			Status:    "success",
			Success:   true,
			TickCount: 5,
			DurationMs: 120,
		})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	tree := NewSequenceNode("root",
		NewActionNode("step1", "noop", nil),
	)
	bt := NewBehaviorTree(tree, rt)

	result, err := bt.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success to be true")
	}
	if result.Status != "success" {
		t.Errorf("expected status success, got %s", result.Status)
	}
	if result.TickCount != 5 {
		t.Errorf("expected tick_count 5, got %d", result.TickCount)
	}
	if result.DurationMs != 120 {
		t.Errorf("expected duration_ms 120, got %d", result.DurationMs)
	}
}

func TestBTreeRunWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["max_ticks"] == nil {
			t.Error("expected max_ticks in request body")
		}
		if int(body["max_ticks"].(float64)) != 100 {
			t.Errorf("expected max_ticks 100, got %v", body["max_ticks"])
		}
		if body["timeout_ms"] == nil {
			t.Error("expected timeout_ms in request body")
		}
		if int(body["timeout_ms"].(float64)) != 5000 {
			t.Errorf("expected timeout_ms 5000, got %v", body["timeout_ms"])
		}
		if body["context"] == nil {
			t.Error("expected context in request body")
		}
		ctxMap := body["context"].(map[string]interface{})
		if ctxMap["env"] != "test" {
			t.Errorf("expected context env=test, got %v", ctxMap["env"])
		}

		json.NewEncoder(w).Encode(BTreeRunResult{
			Status:          "failure",
			Success:         false,
			TickCount:       100,
			DurationMs:      4500,
			MaxTicksReached: true,
		})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	bt := NewBehaviorTree(NewSequenceNode("root"), rt)

	result, err := bt.Run(context.Background(), &BTreeRunOptions{
		Context:   map[string]interface{}{"env": "test"},
		MaxTicks:  100,
		TimeoutMs: 5000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("expected success to be false")
	}
	if !result.MaxTicksReached {
		t.Error("expected max_ticks_reached to be true")
	}
	if result.TickCount != 100 {
		t.Errorf("expected tick_count 100, got %d", result.TickCount)
	}
}

func TestBTreeDeploy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/btree/deploy" {
			t.Errorf("expected path /v1/btree/deploy, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "patrol_bot" {
			t.Errorf("expected name patrol_bot, got %v", body["name"])
		}
		if body["description"] != "Patrol and report" {
			t.Errorf("expected description 'Patrol and report', got %v", body["description"])
		}
		if body["tree"] == nil {
			t.Error("expected tree field in request body")
		}

		json.NewEncoder(w).Encode(BTreeDeployResult{
			Deployed:    true,
			Name:        "patrol_bot",
			Description: "Patrol and report",
		})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	tree := NewSelectorNode("patrol",
		NewConditionNode("is_safe", "threat_level", "none"),
		NewActionNode("alert", "send_alert", map[string]interface{}{"level": "high"}),
	)
	bt := NewBehaviorTree(tree, rt)

	result, err := bt.Deploy(context.Background(), "patrol_bot", "Patrol and report")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Deployed {
		t.Error("expected deployed to be true")
	}
	if result.Name != "patrol_bot" {
		t.Errorf("expected name patrol_bot, got %s", result.Name)
	}
	if result.Description != "Patrol and report" {
		t.Errorf("expected description 'Patrol and report', got %s", result.Description)
	}
}

func TestBTreeDeployWithoutDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if _, exists := body["description"]; exists {
			t.Error("expected description to be omitted when empty")
		}

		json.NewEncoder(w).Encode(BTreeDeployResult{
			Deployed: true,
			Name:     "simple_bot",
		})
	}))
	defer server.Close()

	rt := NewRuntime(server.URL)
	bt := NewBehaviorTree(NewSequenceNode("root"), rt)

	result, err := bt.Deploy(context.Background(), "simple_bot", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Deployed {
		t.Error("expected deployed to be true")
	}
	if result.Name != "simple_bot" {
		t.Errorf("expected name simple_bot, got %s", result.Name)
	}
}

func TestNewSequenceNode(t *testing.T) {
	child1 := NewActionNode("a1", "tool1", nil)
	child2 := NewActionNode("a2", "tool2", map[string]interface{}{"key": "val"})
	node := NewSequenceNode("seq", child1, child2)

	if node["type"] != "sequence" {
		t.Errorf("expected type sequence, got %v", node["type"])
	}
	if node["name"] != "seq" {
		t.Errorf("expected name seq, got %v", node["name"])
	}
	children := node["children"].([]map[string]interface{})
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if children[0]["name"] != "a1" {
		t.Errorf("expected first child name a1, got %v", children[0]["name"])
	}
	if children[1]["name"] != "a2" {
		t.Errorf("expected second child name a2, got %v", children[1]["name"])
	}
}

func TestNewSelectorNode(t *testing.T) {
	node := NewSelectorNode("sel",
		NewConditionNode("check", "ready", "true"),
		NewActionNode("fallback", "retry", nil),
	)

	if node["type"] != "selector" {
		t.Errorf("expected type selector, got %v", node["type"])
	}
	if node["name"] != "sel" {
		t.Errorf("expected name sel, got %v", node["name"])
	}
	children := node["children"].([]map[string]interface{})
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	if children[0]["type"] != "condition" {
		t.Errorf("expected first child type condition, got %v", children[0]["type"])
	}
	if children[1]["type"] != "action" {
		t.Errorf("expected second child type action, got %v", children[1]["type"])
	}
}

func TestNewActionNode(t *testing.T) {
	node := NewActionNode("do_thing", "my_tool", map[string]interface{}{"param": 42})

	if node["type"] != "action" {
		t.Errorf("expected type action, got %v", node["type"])
	}
	if node["name"] != "do_thing" {
		t.Errorf("expected name do_thing, got %v", node["name"])
	}
	if node["tool"] != "my_tool" {
		t.Errorf("expected tool my_tool, got %v", node["tool"])
	}
	args := node["args"].(map[string]interface{})
	if args["param"] != 42 {
		t.Errorf("expected args param 42, got %v", args["param"])
	}
}

func TestNewConditionNode(t *testing.T) {
	node := NewConditionNode("is_ready", "status", "active")

	if node["type"] != "condition" {
		t.Errorf("expected type condition, got %v", node["type"])
	}
	if node["name"] != "is_ready" {
		t.Errorf("expected name is_ready, got %v", node["name"])
	}
	if node["key"] != "status" {
		t.Errorf("expected key status, got %v", node["key"])
	}
	if node["expected"] != "active" {
		t.Errorf("expected expected active, got %v", node["expected"])
	}
}
