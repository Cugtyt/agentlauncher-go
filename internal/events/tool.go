package events

import "agentlauncher/internal/eventbus"

type ToolCall struct {
	eventbus.BaseEvent
	AgentID    string         `json:"agent_id"`
	ToolName   string         `json:"tool_name"`
	ToolCallID string         `json:"tool_call_id"`
	Arguments  map[string]any `json:"arguments"`
}

type ToolResult struct {
	eventbus.BaseEvent
	AgentID    string `json:"agent_id"`
	ToolName   string `json:"tool_name"`
	ToolCallID string `json:"tool_call_id"`
	Result     string `json:"result"`
}

type ToolsExecRequestEvent struct {
	eventbus.BaseEvent
	AgentID   string     `json:"agent_id"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

type ToolsExecResultsEvent struct {
	eventbus.BaseEvent
	AgentID     string       `json:"agent_id"`
	ToolResults []ToolResult `json:"tool_results"`
}

type ToolRuntimeErrorEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}

type ToolExecStartEvent struct {
	eventbus.BaseEvent
	AgentID    string         `json:"agent_id"`
	ToolCallID string         `json:"tool_call_id"`
	ToolName   string         `json:"tool_name"`
	Arguments  map[string]any `json:"arguments"`
}

type ToolExecFinishEvent struct {
	eventbus.BaseEvent
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	ToolName   string `json:"tool_name"`
	Result     string `json:"result"`
}

type ToolExecErrorEvent struct {
	eventbus.BaseEvent
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	ToolName   string `json:"tool_name"`
	Error      string `json:"error"`
}
