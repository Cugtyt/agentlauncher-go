package llminterface

type Message interface {
	message()
	IsResponse() bool
}

type baseMessage struct{}

func (baseMessage) message()         {}
func (baseMessage) IsResponse() bool { return false }

type UserMessage struct {
	baseMessage
	Content string `json:"content"`
}

type SystemMessage struct {
	baseMessage
	Content string `json:"content"`
}

type AssistantMessage struct {
	baseMessage
	Content string `json:"content"`
}

func (m AssistantMessage) IsResponse() bool { return true }

type ToolCallMessage struct {
	baseMessage
	ToolCallID string         `json:"tool_call_id"`
	ToolName   string         `json:"tool_name"`
	Arguments  map[string]any `json:"arguments"`
}

func (m ToolCallMessage) IsResponse() bool { return true }

type ToolResultMessage struct {
	baseMessage
	ToolCallID string `json:"tool_call_id"`
	ToolName   string `json:"tool_name"`
	Result     string `json:"result"`
}

type RequestToolList []ToolSchema
type RequestMessageList []Message
type ResponseMessageList []Message
type MessageList []Message
