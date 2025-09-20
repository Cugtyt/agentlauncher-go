package runtimes

import (
	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/events"
	"agentlauncher/internal/llminterface"
	"context"
	"sync"
)

type MessageRuntime struct {
	History                  map[string][]llminterface.Message `json:"history"`
	eventBus                 *eventbus.EventBus
	response_message_handler func(llminterface.ResponseMessageList) llminterface.ResponseMessageList
	conversation_handler     func(llminterface.MessageList) llminterface.MessageList
	mu                       sync.RWMutex
}

func NewMessageRuntime(
	eventBus *eventbus.EventBus,
) *MessageRuntime {
	messageRuntime := &MessageRuntime{
		History:  make(map[string][]llminterface.Message),
		eventBus: eventBus,
	}
	eventbus.Subscribe(eventBus, messageRuntime.HandleLLMResponseEvent)
	eventbus.Subscribe(eventBus, messageRuntime.HandleTaskCreateEvent)
	eventbus.Subscribe(eventBus, messageRuntime.HandleToolsExecResults)
	eventbus.Subscribe(eventBus, messageRuntime.HandleMessagesAddEvent)
	eventbus.Subscribe(eventBus, messageRuntime.HandleAgentLauncherShutdownEvent)
	eventbus.Subscribe(eventBus, messageRuntime.HandleTaskFinishEvent)
	return messageRuntime
}

func (r *MessageRuntime) WithResponseMessageHandler(handler func(llminterface.ResponseMessageList) llminterface.ResponseMessageList) *MessageRuntime {
	r.response_message_handler = handler
	return r
}

func (r *MessageRuntime) WithConversationHandler(handler func(llminterface.MessageList) llminterface.MessageList) *MessageRuntime {
	r.conversation_handler = handler
	return r
}

func (r *MessageRuntime) HandleLLMResponseEvent(ctx context.Context, e events.LLMResponseEvent) {
	if !IsPrimaryAgent(e.AgentID) {
		return
	}

	responseMessages := e.Response
	if r.response_message_handler != nil {
		responseMessages = r.response_message_handler(e.Response)
	}

	r.mu.Lock()
	if _, exists := r.History[e.AgentID]; !exists {
		panic("History for primary agent " + e.AgentID + " does not exist")
	}
	r.History[e.AgentID] = append(r.History[e.AgentID], responseMessages...)
	r.mu.Unlock()
	r.eventBus.Emit(events.MessagesAddEvent{
		Messages: llminterface.MessageList(responseMessages),
		AgentID:  e.AgentID,
	})
}

func (r *MessageRuntime) HandleTaskCreateEvent(ctx context.Context, e events.TaskCreateEvent) {
	if !IsPrimaryAgent(e.AgentID) {
		return
	}
	r.mu.Lock()
	if _, exists := r.History[e.AgentID]; !exists {
		r.History[e.AgentID] = []llminterface.Message{}
	}
	if e.Conversation != nil {
		r.History[e.AgentID] = append(r.History[e.AgentID], e.Conversation...)
	}
	r.History[e.AgentID] = append(r.History[e.AgentID], llminterface.UserMessage{Content: e.Task})
	r.mu.Unlock()
	r.eventBus.Emit(events.MessagesAddEvent{
		Messages: llminterface.MessageList{llminterface.UserMessage{Content: e.Task}},
		AgentID:  e.AgentID,
	})
}

func (r *MessageRuntime) HandleToolsExecResults(ctx context.Context, e events.ToolsExecResultsEvent) {
	if !IsPrimaryAgent(e.AgentID) {
		return
	}
	toolMessages := []llminterface.Message{}
	for _, result := range e.ToolResults {
		toolMessages = append(toolMessages, llminterface.ToolResultMessage{ToolCallID: result.ToolCallID, ToolName: result.ToolName, Result: result.Result})
	}
	r.mu.Lock()
	if _, exists := r.History[e.AgentID]; !exists {
		panic("History for primary agent " + e.AgentID + " does not exist")
	}
	r.History[e.AgentID] = append(r.History[e.AgentID], toolMessages...)
	r.mu.Unlock()
	r.eventBus.Emit(events.MessagesAddEvent{
		Messages: toolMessages,
		AgentID:  e.AgentID,
	})
}

func (r *MessageRuntime) HandleMessagesAddEvent(ctx context.Context, e events.MessagesAddEvent) {
	if !IsPrimaryAgent(e.AgentID) {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.History[e.AgentID]; !exists {
		panic("History for primary agent " + e.AgentID + " does not exist")
	}
	r.History[e.AgentID] = append(r.History[e.AgentID], e.Messages...)
}

func (r *MessageRuntime) HandleAgentLauncherShutdownEvent(ctx context.Context, e events.AgentLauncherShutdownEvent) {
	if !IsPrimaryAgent(e.AgentID) {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.History[e.AgentID]; !exists {
		panic("History for primary agent " + e.AgentID + " does not exist")
	}
	delete(r.History, e.AgentID)
}

func (r *MessageRuntime) HandleTaskFinishEvent(ctx context.Context, e events.TaskFinishEvent) {
	if !IsPrimaryAgent(e.AgentID) {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.History[e.AgentID]; !exists {
		panic("History for primary agent " + e.AgentID + " does not exist")
	}
	delete(r.History, e.AgentID)
}
