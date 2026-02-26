package igris

import (
	"context"
	"fmt"
	"net/http"

	"gopkg.in/yaml.v3"
)

// BehaviorTree represents a behavior tree that can be validated, executed, and deployed via the Runtime.
type BehaviorTree struct {
	tree    map[string]interface{}
	runtime *Runtime
}

// BTreeValidateResult is the response from validating a behavior tree definition.
type BTreeValidateResult struct {
	Valid    bool   `json:"valid"`
	RootType string `json:"root_type,omitempty"`
	RootName string `json:"root_name,omitempty"`
	Error    string `json:"error,omitempty"`
}

// BTreeRunResult is the response from executing a behavior tree.
type BTreeRunResult struct {
	Status           string `json:"status"`
	Success          bool   `json:"success"`
	TickCount        uint64 `json:"tick_count"`
	DurationMs       uint64 `json:"duration_ms"`
	Cancelled        bool   `json:"cancelled"`
	MaxTicksReached  bool   `json:"max_ticks_reached"`
	DeadlineExceeded bool   `json:"deadline_exceeded"`
	Error            string `json:"error,omitempty"`
}

// BTreeRunOptions configures execution parameters for a behavior tree run.
type BTreeRunOptions struct {
	Context   map[string]interface{} `json:"context,omitempty"`
	MaxTicks  int                    `json:"max_ticks,omitempty"`
	TimeoutMs int                    `json:"timeout_ms,omitempty"`
}

// BTreeDeployResult is the response from deploying a behavior tree.
type BTreeDeployResult struct {
	Deployed    bool   `json:"deployed"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// NewBehaviorTree creates a new BehaviorTree bound to the given runtime.
func NewBehaviorTree(tree map[string]interface{}, runtime *Runtime) *BehaviorTree {
	return &BehaviorTree{
		tree:    tree,
		runtime: runtime,
	}
}

// NewBehaviorTreeFromYAML creates a new BehaviorTree from a YAML string.
func NewBehaviorTreeFromYAML(yamlStr string, runtime *Runtime) (*BehaviorTree, error) {
	var tree map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &tree); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return &BehaviorTree{
		tree:    tree,
		runtime: runtime,
	}, nil
}

// Validate sends the behavior tree definition to the runtime for validation.
func (bt *BehaviorTree) Validate(ctx context.Context) (*BTreeValidateResult, error) {
	payload := map[string]interface{}{
		"tree": bt.tree,
	}
	var result BTreeValidateResult
	if err := bt.runtime.doLocalRequest(ctx, http.MethodPost, "/v1/btree/validate", payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Run executes the behavior tree on the runtime with optional configuration.
func (bt *BehaviorTree) Run(ctx context.Context, opts *BTreeRunOptions) (*BTreeRunResult, error) {
	payload := map[string]interface{}{
		"tree": bt.tree,
	}
	if opts != nil {
		if opts.Context != nil {
			payload["context"] = opts.Context
		}
		if opts.MaxTicks > 0 {
			payload["max_ticks"] = opts.MaxTicks
		}
		if opts.TimeoutMs > 0 {
			payload["timeout_ms"] = opts.TimeoutMs
		}
	}
	var result BTreeRunResult
	if err := bt.runtime.doLocalRequest(ctx, http.MethodPost, "/v1/btree/run", payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Deploy registers the behavior tree under a name on the runtime.
func (bt *BehaviorTree) Deploy(ctx context.Context, name string, description string) (*BTreeDeployResult, error) {
	payload := map[string]interface{}{
		"tree": bt.tree,
		"name": name,
	}
	if description != "" {
		payload["description"] = description
	}
	var result BTreeDeployResult
	if err := bt.runtime.doLocalRequest(ctx, http.MethodPost, "/v1/btree/deploy", payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// NewSequenceNode builds a sequence composite node definition.
// A sequence ticks children in order and succeeds only if all children succeed.
func NewSequenceNode(name string, children ...map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":     "sequence",
		"name":     name,
		"children": children,
	}
}

// NewSelectorNode builds a selector (fallback) composite node definition.
// A selector ticks children in order and succeeds as soon as one child succeeds.
func NewSelectorNode(name string, children ...map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":     "selector",
		"name":     name,
		"children": children,
	}
}

// NewActionNode builds an action leaf node definition that invokes a tool.
func NewActionNode(name string, tool string, args map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type": "action",
		"name": name,
		"tool": tool,
		"args": args,
	}
}

// NewConditionNode builds a condition leaf node that checks a context key against an expected value.
func NewConditionNode(name string, key string, expected string) map[string]interface{} {
	return map[string]interface{}{
		"type":     "condition",
		"name":     name,
		"key":      key,
		"expected": expected,
	}
}

// NewLLMNode builds an LLM leaf node that invokes a language model with a prompt.
func NewLLMNode(name string, prompt string, model string) map[string]interface{} {
	node := map[string]interface{}{
		"type":   "llm_node",
		"name":   name,
		"prompt": prompt,
	}
	if model != "" {
		node["model"] = model
	}
	return node
}

// NewFallbackNode is an alias for NewSelectorNode (tries children until one succeeds).
var NewFallbackNode = NewSelectorNode
