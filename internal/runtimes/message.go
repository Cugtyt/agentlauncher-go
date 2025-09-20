package runtimes

import (
	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/events"
	"agentlauncher/internal/llminterface"
	"context"
)

type MessageRuntime struct {
	History                  []llminterface.Message `json:"history"`
	eventBus                 *eventbus.EventBus
	response_message_handler func(llminterface.ResponseMessageList) llminterface.ResponseMessageList
	conversation_handler     func(llminterface.MessageList) llminterface.MessageList
}

func NewMessageRuntime(
	eventBus *eventbus.EventBus,
) *MessageRuntime {
	messageRuntime := &MessageRuntime{
		History:  []llminterface.Message{},
		eventBus: eventBus,
	}
	eventbus.Subscribe(eventBus, messageRuntime.HandleLLMResponseEvent)
	eventbus.Subscribe(eventBus, messageRuntime.HandleTaskCreateEvent)
	eventbus.Subscribe(eventBus, messageRuntime.HandleToolsExecResults)
	eventbus.Subscribe(eventBus, messageRuntime.HandleMessagesAddEvent)
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
	if e.AgentID != AGENT_0_NAME {
		return
	}

	responseMessages := e.Response
	if r.response_message_handler != nil {
		responseMessages = r.response_message_handler(e.Response)
	}

	r.History = append(r.History, responseMessages...)
	r.eventBus.Emit(events.MessagesAddEvent{
		Messages: llminterface.MessageList(responseMessages),
		AgentID:  e.AgentID,
	})
}

func (r *MessageRuntime) HandleTaskCreateEvent(ctx context.Context, e events.TaskCreateEvent) {
	r.History = append(r.History, llminterface.UserMessage{Content: e.Task})
	r.eventBus.Emit(events.MessagesAddEvent{
		Messages: llminterface.MessageList{llminterface.UserMessage{Content: e.Task}},
		AgentID:  AGENT_0_NAME,
	})
}

func (r *MessageRuntime) HandleToolsExecResults(ctx context.Context, e events.ToolsExecResultsEvent) {
	if e.AgentID != AGENT_0_NAME {
		return
	}
	toolMessages := []llminterface.Message{}
	for _, result := range e.ToolResults {
		toolMessages = append(toolMessages, llminterface.ToolResultMessage{ToolCallID: result.ToolCallID, ToolName: result.ToolName, Result: result.Result})
	}
	r.History = append(r.History, toolMessages...)
	r.eventBus.Emit(events.MessagesAddEvent{
		Messages: toolMessages,
		AgentID:  e.AgentID,
	})
}

func (r *MessageRuntime) HandleMessagesAddEvent(ctx context.Context, e events.MessagesAddEvent) {
	if e.AgentID == AGENT_0_NAME && r.conversation_handler != nil {
		r.History = r.conversation_handler(r.History)
	}
}
