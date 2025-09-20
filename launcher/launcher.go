package launcher

import (
	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/events"
	"agentlauncher/internal/llminterface"
	"agentlauncher/internal/runtimes"
	"context"
	"sync"
	"time"
)

type AgentLauncher struct {
	eventBus       *eventbus.EventBus
	systemPrompt   string
	agentRuntime   *runtimes.AgentRuntime
	llmRuntime     *runtimes.LLMRuntime
	toolRuntime    *runtimes.ToolRuntime
	messageRuntime *runtimes.MessageRuntime
	finalResults   map[string]chan string
	primaryAgents  map[string]bool
	mu             sync.RWMutex
	subAgentTool   bool
}

func NewAgentLauncher(mainAgentHandler llminterface.LLMHandler, subAgentHandler llminterface.LLMHandler) *AgentLauncher {
	eb := eventbus.NewEventBus()
	al := &AgentLauncher{
		eventBus:       eb,
		agentRuntime:   runtimes.NewAgentRuntime(eb),
		llmRuntime:     runtimes.NewLLMRuntime(eb, mainAgentHandler, subAgentHandler),
		toolRuntime:    runtimes.NewToolRuntime(eb),
		messageRuntime: runtimes.NewMessageRuntime(eb),
		systemPrompt:   runtimes.PRIMARY_AGENT_SYSTEM_PROMPT,
		finalResults:   make(map[string]chan string),
		primaryAgents:  make(map[string]bool),
	}

	eventbus.Subscribe(eb, al.HandleTaskFinishEvent)

	return al
}

func (al *AgentLauncher) WithVerboseLevel(level eventbus.VerboseLevel) *AgentLauncher {
	al.eventBus.WithVerboseLevel(level)
	return al
}

func (al *AgentLauncher) WithSystemPrompt(prompt string) *AgentLauncher {
	al.systemPrompt = prompt
	return al
}

func (al *AgentLauncher) WithResponseMessageHandler(handler func(llminterface.ResponseMessageList) llminterface.ResponseMessageList) *AgentLauncher {
	al.messageRuntime.WithResponseMessageHandler(handler)
	return al
}

func (al *AgentLauncher) WithConversationHandler(handler func(llminterface.MessageList) llminterface.MessageList) *AgentLauncher {
	al.messageRuntime.WithConversationHandler(handler)
	return al
}

func (al *AgentLauncher) WithTool(name, description string, fn any, params []llminterface.ToolParamSchema) *AgentLauncher {
	al.toolRuntime.Register(name, description, fn, params)
	return al
}

func (al *AgentLauncher) DisableSubAgentTool() *AgentLauncher {
	al.toolRuntime.DisableSubAgentTool()
	return al
}

func SubscribeEvent[T eventbus.Event](al *AgentLauncher, handler func(context.Context, T)) *AgentLauncher {
	eventbus.Subscribe(al.eventBus, handler)
	return al
}

func (al *AgentLauncher) HandleTaskFinishEvent(ctx context.Context, e events.TaskFinishEvent) {
	al.mu.RLock()
	ch, exists := al.finalResults[e.AgentID]
	al.mu.RUnlock()

	if exists {
		select {
		case ch <- e.Result:
		default:
		}
	}
}

func (al *AgentLauncher) Run(task string, history []llminterface.Message) string {
	al.mu.Lock()
	agentID := runtimes.GeneratePrimaryAgentID(len(al.primaryAgents))
	al.primaryAgents[agentID] = true
	al.toolRuntime.SetupSubAgentTool()
	al.finalResults[agentID] = make(chan string, 1)
	al.mu.Unlock()

	tool_names := al.toolRuntime.GetToolNames()
	al.eventBus.Emit(events.TaskCreateEvent{
		AgentID:      agentID,
		Task:         task,
		Conversation: history,
		SystemPrompt: al.systemPrompt,
		ToolSchemas:  al.toolRuntime.GetToolSchemas(tool_names),
	})
	select {
	case result := <-al.finalResults[agentID]:
		al.mu.Lock()
		delete(al.finalResults, agentID)
		al.mu.Unlock()
		return result
	case <-time.After(30 * time.Minute):
		return "Task timed out"
	}
}

func (al *AgentLauncher) Close() {
	al.eventBus.Shutdown(context.Background())
}
