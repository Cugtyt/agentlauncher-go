package events

import (
	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/llminterface"
)

type TaskCreateEvent struct {
	eventbus.BaseEvent
	AgentID      string                    `json:"agent_id"`
	Task         string                    `json:"task"`
	ToolSchemas  []llminterface.ToolSchema `json:"tool_schemas"`
	SystemPrompt string                    `json:"system_prompt"`
	Conversation []llminterface.Message    `json:"conversation"`
}

type TaskFinishEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
	Result  string `json:"result"`
}
