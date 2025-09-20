package main

import (
	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/events"
	"agentlauncher/internal/llminterface"
	"agentlauncher/launcher"
	"context"
	"encoding/json"
	"fmt"
)

func main() {
	agentLauncher := launcher.NewAgentLauncher(MainAgentLLMHandler, MainAgentLLMHandler).WithVerboseLevel(eventbus.SILENT)
	RegisterTools(agentLauncher)
	launcher.SubscribeEvent(agentLauncher, func(ctx context.Context, event events.MessagesAddEvent) {
		fmt.Println("[", event.AgentID, "] Messages added:")
		for _, msg := range event.Messages {
			switch msg := msg.(type) {
			case llminterface.UserMessage:
				fmt.Println("User:", msg.Content)
			case llminterface.AssistantMessage:
				fmt.Println("Assistant:", msg.Content)
			case llminterface.ToolResultMessage:
				fmt.Println("Tool Result:", msg.ToolName, "(", msg.ToolCallID, ") ", msg.Result)
			case llminterface.ToolCallMessage:
				args, _ := json.Marshal(msg.Arguments)
				fmt.Println("Tool Call:", msg.ToolName, "(", msg.ToolCallID, ") with args", string(args))
			}
		}
	})
	response := agentLauncher.Run(`You are to help me organize a virtual conference. Please:
1. Find three suitable dates in the next month for the event.
2. Research and suggest two keynote speakers in AI.
3. Prepare a draft agenda with at least five sessions.
4. List three online platforms suitable for hosting the conference.
5. Estimate a budget for the event including speaker fees, platform costs and marketing.
6. Draft an invitation email for potential attendees.
7. Summarize all findings and provide a recommended plan of action.
Each step may require different tools or information sources. Provide a clear summary.`)
	fmt.Println("Final response:", response)
}
