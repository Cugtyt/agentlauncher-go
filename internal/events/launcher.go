package events

import "agentlauncher/internal/eventbus"

type AgentLauncherRunEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
	Task    string `json:"task"`
}

type AgentLauncherStopEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
	Task    string `json:"task"`
}

type AgentLauncherShutdownEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
}

type AgentLauncherErrorEvent struct {
	eventbus.BaseEvent
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}
