package llminterface

type ToolParamSchema struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Required    bool           `json:"required"`
	Items       map[string]any `json:"items,omitempty"`
}

type ToolSchema struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  []ToolParamSchema `json:"parameters"`
}
