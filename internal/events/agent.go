package events

import (
	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/llminterface"
)

type AgentCreateEvent struct {
	eventbus.BaseEvent
	AgentID      string                    `json:"agent_id"`
	Task         string                    `json:"task"`
	ToolSchemas  []llminterface.ToolSchema `json:"tool_schemas"`
	Conversation []llminterface.Message    `json:"conversation"`
	SystemPrompt string                    `json:"system_prompt"`
}

type AgentStartEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
}

type AgentFinishEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
	Result  string `json:"result"`
}

type AgentRuntimeErrorEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}

type AgentDeletedEvent struct {
	AgentID string `json:"agent_id"`
	eventbus.BaseEvent
}
