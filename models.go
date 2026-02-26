package igris

import "context"

// ModelManager provides model management operations via the local Runtime (BYOM).
type ModelManager struct {
	runtime *Runtime
}

// NewModelManager creates a ModelManager backed by a Runtime.
func NewModelManager(runtime *Runtime) *ModelManager {
	return &ModelManager{runtime: runtime}
}

// UploadModel loads a GGUF model into the local runtime.
func (m *ModelManager) UploadModel(ctx context.Context, modelPath, modelID string) (map[string]interface{}, error) {
	return m.runtime.LoadModel(ctx, modelPath, modelID)
}

// ListLocalModels lists models available on the local runtime.
func (m *ModelManager) ListLocalModels(ctx context.Context) ([]map[string]interface{}, error) {
	return m.runtime.ListModels(ctx)
}

// SetActiveModel hot-swaps to a different model on the runtime.
func (m *ModelManager) SetActiveModel(ctx context.Context, modelID string) (map[string]interface{}, error) {
	return m.runtime.SwapModel(ctx, modelID)
}
