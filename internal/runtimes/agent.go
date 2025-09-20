package runtimes

import (
	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/events"
	"context"
	"sync"
)

type AgentRuntime struct {
	Agents   map[string]*Agent `json:"agents"`
	eventBus *eventbus.EventBus
	mu       sync.RWMutex
}

func NewAgentRuntime(eb *eventbus.EventBus) *AgentRuntime {
	agentRuntime := &AgentRuntime{
		Agents:   make(map[string]*Agent),
		eventBus: eb,
	}

	eventbus.Subscribe(eb, agentRuntime.HandleTaskCreateEvent)
	eventbus.Subscribe(eb, agentRuntime.HandleAgentCreateEvent)
	eventbus.Subscribe(eb, agentRuntime.HandleLLMResponseEvent)
	eventbus.Subscribe(eb, agentRuntime.HandleToolsExecResults)
	eventbus.Subscribe(eb, agentRuntime.HandleAgentFinishEvent)
	eventbus.Subscribe(eb, agentRuntime.HandleAgentRuntimeErrorEvent)
	eventbus.Subscribe(eb, agentRuntime.HandleAgentLauncherShutdownEvent)
	eventbus.Subscribe(eb, agentRuntime.HandleTaskFinishEvent)

	return agentRuntime
}

func (r *AgentRuntime) GetAgent(agentID string) (*Agent, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	agent, exists := r.Agents[agentID]
	return agent, exists
}

func (r *AgentRuntime) HandleTaskCreateEvent(ctx context.Context, e events.TaskCreateEvent) {
	r.eventBus.Emit(events.AgentCreateEvent{
		AgentID:      AGENT_0_NAME,
		Task:         e.Task,
		Conversation: e.Conversation,
		SystemPrompt: func() string {
			if e.SystemPrompt != "" {
				return e.SystemPrompt
			}
			return AGENT_0_SYSTEM_PROMPT
		}(),
		ToolSchemas: e.ToolSchemas,
	})
}

func (r *AgentRuntime) HandleAgentCreateEvent(ctx context.Context, e events.AgentCreateEvent) {
	if _, exists := r.GetAgent(e.AgentID); exists {
		r.eventBus.Emit(events.AgentRuntimeErrorEvent{
			AgentID: e.AgentID,
			Error:   "Agent with this ID already exists",
		})
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Agents[e.AgentID] = NewAgent(
		e.AgentID,
		e.Task,
		e.ToolSchemas,
		r.eventBus,
		e.SystemPrompt,
	)
	go r.Agents[e.AgentID].Start()
}

func (r *AgentRuntime) HandleLLMResponseEvent(ctx context.Context, e events.LLMResponseEvent) {
	if agent, exists := r.GetAgent(e.AgentID); !exists {
		r.eventBus.Emit(events.AgentRuntimeErrorEvent{
			AgentID: e.AgentID,
			Error:   "Agent not found",
		})
	} else {
		go agent.HandleLLMResponse(e.Response)
	}
}

func (r *AgentRuntime) HandleToolsExecResults(ctx context.Context, e events.ToolsExecResultsEvent) {
	if agent, exists := r.GetAgent(e.AgentID); !exists {
		r.eventBus.Emit(events.AgentRuntimeErrorEvent{
			AgentID: e.AgentID,
			Error:   "Agent not found",
		})
	} else {
		go agent.HandleToolsExecResults(e.ToolResults)
	}
}

func (r *AgentRuntime) HandleAgentFinishEvent(ctx context.Context, e events.AgentFinishEvent) {
	if _, exists := r.GetAgent(e.AgentID); !exists {
		r.eventBus.Emit(events.AgentRuntimeErrorEvent{
			AgentID: e.AgentID,
			Error:   "Agent not found",
		})
	} else {
		if e.AgentID != AGENT_0_NAME {
			r.mu.Lock()
			delete(r.Agents, e.AgentID)
			r.mu.Unlock()
			r.eventBus.Emit(events.AgentDeletedEvent{AgentID: e.AgentID})
		} else {
			r.eventBus.Emit(events.TaskFinishEvent{
				AgentID: e.AgentID,
				Result:  e.Result,
			})
		}
	}
}

func (r *AgentRuntime) HandleAgentRuntimeErrorEvent(ctx context.Context, e events.AgentRuntimeErrorEvent) {
	if _, exists := r.GetAgent(e.AgentID); exists {
		r.mu.Lock()
		delete(r.Agents, e.AgentID)
		r.mu.Unlock()
		r.eventBus.Emit(events.AgentDeletedEvent{AgentID: e.AgentID})
	}
	r.eventBus.Emit(events.TaskFinishEvent{
		AgentID: e.AgentID,
		Result:  "Error: " + e.Error,
	})
}

func (r *AgentRuntime) HandleAgentLauncherShutdownEvent(ctx context.Context, e events.AgentLauncherShutdownEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for agentID := range r.Agents {
		delete(r.Agents, agentID)
		r.eventBus.Emit(events.AgentDeletedEvent{AgentID: agentID})
	}
}

func (r *AgentRuntime) HandleTaskFinishEvent(ctx context.Context, e events.TaskFinishEvent) {
	if _, exists := r.GetAgent(e.AgentID); exists && e.AgentID == AGENT_0_NAME {
		r.mu.Lock()
		delete(r.Agents, e.AgentID)
		r.mu.Unlock()
		r.eventBus.Emit(events.AgentDeletedEvent{AgentID: e.AgentID})
	}
}
