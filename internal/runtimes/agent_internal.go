package runtimes

import (
	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/events"
	"agentlauncher/internal/llminterface"
)

type Agent struct {
	AgentID      string                    `json:"agent_id"`
	Task         string                    `json:"task"`
	Conversation []llminterface.Message    `json:"conversation"`
	SystemPrompt string                    `json:"system_prompt"`
	ToolSchemas  []llminterface.ToolSchema `json:"tool_schemas"`
	EventBus     *eventbus.EventBus
}

func NewAgent(
	agentID string,
	task string,
	toolSchemas []llminterface.ToolSchema,
	eventBus *eventbus.EventBus,
	systemPrompt string,
) *Agent {
	return &Agent{
		AgentID:      agentID,
		Task:         task,
		Conversation: []llminterface.Message{},
		SystemPrompt: systemPrompt,
		ToolSchemas:  toolSchemas,
		EventBus:     eventBus,
	}
}

func (a *Agent) Start() {
	a.EventBus.Emit(events.AgentStartEvent{AgentID: a.AgentID})
	a.Conversation = append(a.Conversation, llminterface.UserMessage{Content: a.Task})
	messageList := []llminterface.Message{}
	if a.SystemPrompt != "" {
		messageList = append(messageList, llminterface.SystemMessage{Content: a.SystemPrompt})
	}
	messageList = append(messageList, a.Conversation...)
	a.EventBus.Emit(events.LLMRequestEvent{
		AgentID:     a.AgentID,
		Messages:    messageList,
		ToolSchemas: a.ToolSchemas,
	})
}

func (a *Agent) HandleLLMResponse(response llminterface.ResponseMessageList) {
	a.Conversation = append(a.Conversation, response...)
	toolCalls := []events.ToolCall{}
	for _, msg := range response {
		if toolCallMsg, ok := msg.(llminterface.ToolCallMessage); ok {
			toolCalls = append(toolCalls, events.ToolCall{
				ToolCallID: toolCallMsg.ToolCallID,
				ToolName:   toolCallMsg.ToolName,
				Arguments:  toolCallMsg.Arguments,
			})
		}
	}
	if len(toolCalls) == 0 {
		assistantContents := ""
		for _, msg := range response {
			if assistantMsg, ok := msg.(llminterface.AssistantMessage); ok {
				assistantContents += "\n" + assistantMsg.Content
			}
		}
		a.EventBus.Emit(events.AgentFinishEvent{
			AgentID: a.AgentID,
			Result:  assistantContents,
		})
	} else {
		a.EventBus.Emit(events.ToolsExecRequestEvent{
			AgentID:   a.AgentID,
			ToolCalls: toolCalls,
		})
	}
}

func (a *Agent) HandleToolsExecResults(toolResults []events.ToolResult) {
	for _, result := range toolResults {
		a.Conversation = append(a.Conversation, llminterface.ToolResultMessage{
			ToolCallID: result.ToolCallID,
			ToolName:   result.ToolName,
			Result:     result.Result,
		})
	}
	messageList := []llminterface.Message{}
	if a.SystemPrompt != "" {
		messageList = append(messageList, llminterface.SystemMessage{Content: a.SystemPrompt})
	}
	messageList = append(messageList, a.Conversation...)
	a.EventBus.Emit(events.LLMRequestEvent{
		AgentID:     a.AgentID,
		Messages:    messageList,
		ToolSchemas: a.ToolSchemas,
	})
}
