package runtimes

import (
	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/events"
	"agentlauncher/internal/llminterface"
	"context"
)

type LLMRuntime struct {
	eventBus               *eventbus.EventBus
	main_agent_llm_handler llminterface.LLMHandler
	sub_agent_llm_handler  llminterface.LLMHandler
}

func NewLLMRuntime(eventBus *eventbus.EventBus, mainAgentHandler llminterface.LLMHandler, subAgentHandler llminterface.LLMHandler) *LLMRuntime {
	llmRuntime := &LLMRuntime{
		eventBus:               eventBus,
		main_agent_llm_handler: mainAgentHandler,
		sub_agent_llm_handler:  subAgentHandler,
	}
	eventbus.Subscribe(eventBus, llmRuntime.HandleLLMRequestEvent)
	eventbus.Subscribe(eventBus, llmRuntime.HandleLLMRuntimeErrorEvent)
	return llmRuntime
}

func (r *LLMRuntime) HandleLLMRequestEvent(ctx context.Context, event events.LLMRequestEvent) {
	var handler llminterface.LLMHandler
	if event.AgentID == AGENT_0_NAME {
		handler = r.main_agent_llm_handler
	} else {
		handler = r.sub_agent_llm_handler
	}

	if handler == nil {
		r.eventBus.Emit(events.LLMRuntimeErrorEvent{
			AgentID:      event.AgentID,
			Error:        "No LLM handler configured",
			RequestEvent: event,
		})
		return
	}

	response := handler(event.Messages, event.ToolSchemas, event.AgentID, r.eventBus)
	r.eventBus.Emit(events.LLMResponseEvent{
		AgentID:      event.AgentID,
		RequestEvent: event,
		Response:     response,
	})
}

func (r *LLMRuntime) HandleLLMRuntimeErrorEvent(ctx context.Context, event events.LLMRuntimeErrorEvent) {
	if event.RequestEvent.RetryCount < 5 {
		r.eventBus.Emit(events.LLMRequestEvent{
			AgentID:     event.AgentID,
			Messages:    event.RequestEvent.Messages,
			ToolSchemas: event.RequestEvent.ToolSchemas,
			RetryCount:  event.RequestEvent.RetryCount + 1,
		})
	} else {
		response := []llminterface.Message{
			llminterface.AssistantMessage{Content: "Runtime error: " + event.Error},
		}
		r.eventBus.Emit(events.LLMResponseEvent{
			AgentID:      event.AgentID,
			RequestEvent: event.RequestEvent,
			Response:     response,
		})
	}
}
