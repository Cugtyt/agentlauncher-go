package events

import (
	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/llminterface"
)

type MessagesAddEvent struct {
	eventbus.BaseEvent
	AgentID  string                   `json:"agent_id"`
	Messages llminterface.MessageList `json:"messages"`
}

type MessageStartStreamingEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
}

type MessageDeltaStreamingEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
	Delta   string `json:"delta"`
}

type MessageDoneStreamingEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
	Message string `json:"message"`
}

type MessageErrorStreamingEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}

type ToolCallNameStreamingEvent struct {
	eventbus.BaseEvent
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	ToolName   string `json:"tool_name"`
}

type ToolCallArgumentsStartStreamingEvent struct {
	eventbus.BaseEvent
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
}

type ToolCallArgumentsDeltaStreamingEvent struct {
	eventbus.BaseEvent
	AgentID        string `json:"agent_id"`
	ToolCallID     string `json:"tool_call_id"`
	ArgumentsDelta string `json:"arguments_delta"`
}

type ToolCallArgumentsDoneStreamingEvent struct {
	eventbus.BaseEvent
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	Arguments  string `json:"arguments"`
}

type ToolCallArgumentsErrorStreamingEvent struct {
	eventbus.BaseEvent
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	Error      string `json:"error"`
}
