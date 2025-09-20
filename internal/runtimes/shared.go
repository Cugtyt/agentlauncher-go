package runtimes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const (
	PRIMARY_AGENT_PREFIX        string = "agent"
	CREATE_SUB_AGENT_TOOL_NAME  string = "create_sub_agent"
	PRIMARY_AGENT_SYSTEM_PROMPT string = `Your primary role is to wisely delegate tasks
by creating sub-agents whenever a task requires multiple steps or tools.
Try to avoid creating sub-agents for tasks that only require a single step.
Direct execution is 10x more costly than delegation and increases the workload.

For example, if a task requires 5 steps,
- do it yourself could cost 5 llm calls,
- delegating to 1 sub-agent could cost 1 llm call (to create the sub-agent)
+ 5 slm calls (for the sub-agent to complete the task),
but the sub-agent slm calls are 10x cheaper,
so the total cost is 1 + 5/10 = 1.5 llm calls, 0.3x of doing it yourself.

You may launch up to 3 sub-agents at once,
and should run them in parallel whenever possible.
Sub-agents lack access to your task or conversation history,
so always provide complete context and instructions.
Sub-agents will be deleted after returning their results,
so conversation is not available for back-and-forth interactions.
Describe your efficient delegation strategy before creating sub-agents.
Organize results for easy understanding, you don't need to report how you delegated.
`
)

func GeneratePrimaryAgentID(index int) string {
	return fmt.Sprintf("%s%d", PRIMARY_AGENT_PREFIX, index)
}

func GenerateSubAgentID(primaryAgentID string) string {
	return fmt.Sprintf("%s_%s", primaryAgentID, uuid.New().String())
}

func GetPrimaryAgentIDFromSubAgentID(subAgentID string) (string, error) {
	var primaryAgentID string
	n, err := fmt.Sscanf(subAgentID, PRIMARY_AGENT_PREFIX+"%d_", &primaryAgentID)
	if n != 1 || err != nil {
		return "", fmt.Errorf("invalid sub-agent ID format")
	}
	return primaryAgentID, nil
}

func IsPrimaryAgent(agentID string) bool {
    if !strings.HasPrefix(agentID, PRIMARY_AGENT_PREFIX) {
        return false
    }
    
    suffix := agentID[len(PRIMARY_AGENT_PREFIX):]
    index, err := strconv.Atoi(suffix)
    return err == nil && index >= 0
}
