package llminterface

import (
	"agentlauncher/internal/eventbus"
)

type LLMHandler func(messages RequestMessageList, tools RequestToolList, agentid string, eventbus *eventbus.EventBus) ResponseMessageList
