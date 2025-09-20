package events

import (
	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/llminterface"
)

type LLMRequestEvent struct {
	eventbus.BaseEvent
	AgentID     string                    `json:"agent_id"`
	Messages    []llminterface.Message    `json:"messages"`
	ToolSchemas []llminterface.ToolSchema `json:"tool_schemas"`
	RetryCount  int                       `json:"retry_count"`
}

type LLMResponseEvent struct {
	eventbus.BaseEvent
	AgentID      string                 `json:"agent_id"`
	RequestEvent LLMRequestEvent        `json:"request_event"`
	Response     []llminterface.Message `json:"response"`
}

type LLMRuntimeErrorEvent struct {
	eventbus.BaseEvent
	AgentID      string          `json:"agent_id"`
	Error        string          `json:"error"`
	RequestEvent LLMRequestEvent `json:"request_event"`
}
