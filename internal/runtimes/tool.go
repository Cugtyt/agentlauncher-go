package runtimes

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/events"
	"agentlauncher/internal/llminterface"

	"github.com/google/uuid"
)

type Tool struct {
	llminterface.ToolSchema
	Function any
}

type ToolRuntime struct {
	eventBus *eventbus.EventBus
	tools    map[string]*Tool
}

func NewToolRuntime(eventBus *eventbus.EventBus) *ToolRuntime {
	toolRuntime := &ToolRuntime{
		eventBus: eventBus,
		tools:    make(map[string]*Tool),
	}
	eventbus.Subscribe(eventBus, toolRuntime.handleToolsExecRequest)
	return toolRuntime
}

func (tr *ToolRuntime) Register(name, description string, fn any, params []llminterface.ToolParamSchema) {
	if !isValidToolFunction(fn) {
		panic(fmt.Sprintf("invalid tool function signature for %s", name))
	}

	fnType := reflect.TypeOf(fn)
	expectedParams := fnType.NumIn() - 1
	if len(params) != expectedParams {
		panic(fmt.Sprintf("tool '%s' expects %d parameters but got %d", name, expectedParams, len(params)))
	}

	tool := &Tool{
		ToolSchema: llminterface.ToolSchema{
			Name:        name,
			Description: description,
			Parameters:  params,
		},
		Function: fn,
	}

	if _, exists := tr.tools[name]; exists {
		panic(fmt.Sprintf("tool '%s' is already registered", name))
	}

	tr.tools[name] = tool
}

func isValidToolFunction(fn any) bool {
	fnType := reflect.TypeOf(fn)

	if fnType.Kind() != reflect.Func {
		return false
	}

	if fnType.NumIn() < 1 {
		return false
	}

	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !fnType.In(0).Implements(contextType) {
		return false
	}

	if fnType.NumOut() != 2 {
		return false
	}

	errorType := reflect.TypeOf((*error)(nil)).Elem()
	return fnType.Out(1).Implements(errorType)
}

func (tr *ToolRuntime) handleToolsExecRequest(ctx context.Context, event events.ToolsExecRequestEvent) {
	missingTools := make([]string, 0)
	for _, toolCall := range event.ToolCalls {
		if _, exists := tr.tools[toolCall.ToolName]; !exists {
			missingTools = append(missingTools, toolCall.ToolName)
		}
	}

	if len(missingTools) > 0 {
		tr.eventBus.Emit(events.ToolRuntimeErrorEvent{
			AgentID: event.AgentID,
			Error:   fmt.Sprintf("Missing tools: %v", missingTools),
		})
		return
	}

	results := make([]events.ToolResult, len(event.ToolCalls))
	resultsChan := make(chan struct {
		index  int
		result events.ToolResult
	}, len(event.ToolCalls))

	for i, toolCall := range event.ToolCalls {
		go func(index int, tc events.ToolCall) {
			result, err := tr.toolExec(ctx, tc.ToolName, tc.Arguments, event.AgentID, tc.ToolCallID)
			toolResult := events.ToolResult{
				AgentID:    event.AgentID,
				ToolName:   tc.ToolName,
				ToolCallID: tc.ToolCallID,
			}
			if err != nil {
				toolResult.Result = err.Error()
			} else {
				toolResult.Result = result
			}
			resultsChan <- struct {
				index  int
				result events.ToolResult
			}{index: index, result: toolResult}
		}(i, toolCall)
	}

	for i := 0; i < len(event.ToolCalls); i++ {
		r := <-resultsChan
		results[r.index] = r.result
	}
	close(resultsChan)

	tr.eventBus.Emit(events.ToolsExecResultsEvent{
		AgentID:     event.AgentID,
		ToolResults: results,
	})
}

func (tr *ToolRuntime) toolExec(ctx context.Context, toolName string, arguments map[string]any, agentID, toolCallID string) (string, error) {
	tr.eventBus.Emit(events.ToolExecStartEvent{
		AgentID:    agentID,
		ToolCallID: toolCallID,
		ToolName:   toolName,
		Arguments:  arguments,
	})

	tool, exists := tr.tools[toolName]
	if !exists {
		err := fmt.Errorf("tool '%s' not found", toolName)
		tr.emitErrorEvent(agentID, toolCallID, toolName, err)
		return "", err
	}

	result, err := tr.executeToolFunction(ctx, tool, arguments)

	if err != nil {
		tr.emitErrorEvent(agentID, toolCallID, toolName, err)
		return "", err
	}

	tr.eventBus.Emit(events.ToolExecFinishEvent{
		AgentID:    agentID,
		ToolCallID: toolCallID,
		ToolName:   toolName,
		Result:     result,
	})

	return result, nil
}

func (tr *ToolRuntime) executeToolFunction(ctx context.Context, tool *Tool, arguments map[string]any) (string, error) {
	fnValue := reflect.ValueOf(tool.Function)
	fnType := reflect.TypeOf(tool.Function)

	expectedArgs := fnType.NumIn() - 1 // Exclude context
	if len(tool.Parameters) != expectedArgs {
		return "", fmt.Errorf("tool schema mismatch: expected %d parameters, schema has %d", expectedArgs, len(tool.Parameters))
	}

	args := make([]reflect.Value, fnType.NumIn())
	args[0] = reflect.ValueOf(ctx)

	// Map arguments using parameter names from the schema
	for i := 0; i < len(tool.Parameters); i++ {
		paramSchema := tool.Parameters[i]
		paramType := fnType.In(i + 1) // +1 to skip context

		argValue, exists := arguments[paramSchema.Name]
		if !exists {
			if paramSchema.Required {
				return "", fmt.Errorf("missing required argument: %s", paramSchema.Name)
			}
			// Use zero value for optional parameters
			args[i+1] = reflect.Zero(paramType)
			continue
		}

		convertedValue, err := tr.convertArgument(argValue, paramType)
		if err != nil {
			return "", fmt.Errorf("argument '%s': %w", paramSchema.Name, err)
		}

		args[i+1] = convertedValue
	}

	results := fnValue.Call(args)

	if len(results) != 2 {
		return "", fmt.Errorf("tool function must return (result, error)")
	}

	if !results[1].IsNil() {
		return "", results[1].Interface().(error)
	}

	result := results[0].Interface()
	return tr.resultToString(result)
}

func (tr *ToolRuntime) convertArgument(value interface{}, targetType reflect.Type) (reflect.Value, error) {
	if value == nil {
		if targetType.Kind() == reflect.Pointer {
			return reflect.Zero(targetType), nil
		}
		return reflect.Value{}, fmt.Errorf("nil value for required parameter")
	}

	valueType := reflect.TypeOf(value)

	if valueType == targetType {
		return reflect.ValueOf(value), nil
	}

	if valueType.ConvertibleTo(targetType) {
		return reflect.ValueOf(value).Convert(targetType), nil
	}

	switch targetType.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		if num, ok := value.(float64); ok {
			return reflect.ValueOf(int64(num)).Convert(targetType), nil
		}
	case reflect.Float32, reflect.Float64:
		if num, ok := value.(float64); ok {
			return reflect.ValueOf(num).Convert(targetType), nil
		}
	case reflect.String:
		return reflect.ValueOf(fmt.Sprintf("%v", value)), nil
	case reflect.Bool:
		if b, ok := value.(bool); ok {
			return reflect.ValueOf(b), nil
		}
	case reflect.Slice:
		if arr, ok := value.([]any); ok {
			sliceValue := reflect.MakeSlice(targetType, len(arr), len(arr))
			elemType := targetType.Elem()
			for i, item := range arr {
				convertedItem, err := tr.convertArgument(item, elemType)
				if err != nil {
					return reflect.Value{}, fmt.Errorf("array element %d: %w", i, err)
				}
				sliceValue.Index(i).Set(convertedItem)
			}
			return sliceValue, nil
		}
	}
	

	return reflect.Value{}, fmt.Errorf("cannot convert %T to %v", value, targetType)
}

func (tr *ToolRuntime) resultToString(result any) (string, error) {
	switch v := result.(type) {
	case string:
		return v, nil
	case nil:
		return "", nil
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}
		return string(data), nil
	}
}

func (tr *ToolRuntime) Setup() {
	tr.Register("create_sub_agent",
		"Create a sub-agent to handle a specific task",
		tr.createSubAgentTool,
		[]llminterface.ToolParamSchema{
			{
				Type:        "string",
				Name:        "task",
				Description: "Task for the sub-agent to accomplish",
				Required:    true,
			},
			{
				Name:        "toolNameList",
				Type:        "array",
				Description: "List of tool names the sub-agent can use",
				Required:    true,
				Items: map[string]any{
					"type": "string",
				},
			},
		})
}

func (tr *ToolRuntime) createSubAgentTool(ctx context.Context, task string, toolNameList []string) (string, error) {
	agentID := uuid.New().String()
	resultChan := make(chan string, 1)

	eventbus.Subscribe(tr.eventBus, func(eventCtx context.Context, event events.AgentFinishEvent) {
		if event.AgentID == agentID {
			select {
			case resultChan <- event.Result:
			default:
			}
			close(resultChan)
		}
	})

	tr.eventBus.Emit(events.AgentCreateEvent{
		AgentID:     agentID,
		Task:        task,
		ToolSchemas: tr.getToolSchemas(toolNameList),
	})

	select {
	case result := <-resultChan:
		return result, nil
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(5 * time.Minute):
		return "", fmt.Errorf("sub-agent timeout")
	}
}

func (tr *ToolRuntime) getToolSchemas(toolNames []string) []llminterface.ToolSchema {
	schemas := make([]llminterface.ToolSchema, 0, len(toolNames))

	for _, toolName := range toolNames {
		if tool, exists := tr.tools[toolName]; exists {
			schemas = append(schemas, llminterface.ToolSchema{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			})
		}
	}

	return schemas
}

func (tr *ToolRuntime) GetToolSchemas(names []string) []llminterface.ToolSchema {
	schemas := make([]llminterface.ToolSchema, 0, len(names))

	for _, name := range names {
		if tool, exists := tr.tools[name]; exists {
			schemas = append(schemas, tool.ToolSchema)
		}
	}

	return schemas
}

func (tr *ToolRuntime) emitErrorEvent(agentID, toolCallID, toolName string, err error) {
	tr.eventBus.Emit(events.ToolExecErrorEvent{
		AgentID:    agentID,
		ToolCallID: toolCallID,
		ToolName:   toolName,
		Error:      err.Error(),
	})
}

func (tr *ToolRuntime) GetToolNames() []string {
	names := make([]string, 0, len(tr.tools))
	for name := range tr.tools {
		names = append(names, name)
	}
	return names
}
