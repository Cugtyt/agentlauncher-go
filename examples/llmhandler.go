package main

import (
	"agentlauncher/internal/eventbus"
	"agentlauncher/internal/llminterface"
	"context"
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/azure"
	"github.com/openai/openai-go/v2/option"
)

func MainAgentLLMHandler(messages llminterface.RequestMessageList, tools llminterface.RequestToolList, agentid string, eventbus *eventbus.EventBus) llminterface.ResponseMessageList {
	tokenCredential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		panic(err)
	}
	client := openai.NewClient(
		option.WithBaseURL("https://smarttsg-gpt.openai.azure.com/openai/v1/"),
		azure.WithTokenCredential(tokenCredential),
	)
	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, 0)

	// Track if we're building an assistant message with tool calls
	var currentToolCalls []openai.ChatCompletionMessageToolCallUnionParam
	var i int

	for i = range messages {
		msg := messages[i]

		switch m := msg.(type) {
		case llminterface.UserMessage:
			// Flush any pending tool calls
			if len(currentToolCalls) > 0 {
				openaiMessages = append(openaiMessages, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						ToolCalls: currentToolCalls,
					},
				})
				currentToolCalls = nil
			}
			openaiMessages = append(openaiMessages, openai.UserMessage(m.Content))

		case llminterface.AssistantMessage:
			// Flush any pending tool calls
			if len(currentToolCalls) > 0 {
				openaiMessages = append(openaiMessages, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						ToolCalls: currentToolCalls,
					},
				})
				currentToolCalls = nil
			}

			// Check if next messages are tool calls for this assistant message
			hasToolCalls := false
			if i+1 < len(messages) {
                _, hasToolCalls = messages[i+1].(llminterface.ToolCallMessage)
            }

			if hasToolCalls {
				// We'll handle this assistant message with its tool calls
				// Just set the content, tool calls will be added in the next iterations
				// For now, skip adding the message
			} else {
				openaiMessages = append(openaiMessages, openai.AssistantMessage(m.Content))
			}

		case llminterface.SystemMessage:
			// Flush any pending tool calls
			if len(currentToolCalls) > 0 {
				openaiMessages = append(openaiMessages, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						ToolCalls: currentToolCalls,
					},
				})
				currentToolCalls = nil
			}
			openaiMessages = append(openaiMessages, openai.SystemMessage(m.Content))

		case llminterface.ToolCallMessage:
			// Accumulate tool calls
			currentToolCalls = append(currentToolCalls, openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID:   m.ToolCallID,
					Type: "function",
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name: m.ToolName,
						Arguments: func() string {
							argsBytes, _ := json.Marshal(m.Arguments)
							return string(argsBytes)
						}(),
					},
				},
			})

		case llminterface.ToolResultMessage:
			// Flush any pending tool calls before adding tool result
			if len(currentToolCalls) > 0 {
				openaiMessages = append(openaiMessages, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						ToolCalls: currentToolCalls,
					},
				})
				currentToolCalls = nil
			}
			openaiMessages = append(openaiMessages, openai.ToolMessage(m.Result, m.ToolCallID))
		}
	}

	// Flush any remaining tool calls
	if len(currentToolCalls) > 0 {
		openaiMessages = append(openaiMessages, openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				ToolCalls: currentToolCalls,
			},
		})
	}
	openaiTools := make([]openai.ChatCompletionToolUnionParam, len(tools))
	for i, tool := range tools {
		parameters := make(map[string]any)
		required := []string{}
		for _, param := range tool.Parameters {
			parameters[param.Name] = map[string]any{
				"type":        param.Type,
				"description": param.Description,
			}
			if param.Type == "array" && param.Items != nil {
				parameters[param.Name].(map[string]any)["items"] = param.Items
			}
		}

		openaiTools[i] = openai.ChatCompletionToolUnionParam{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Function: openai.FunctionDefinitionParam{
					Name:        tool.Name,
					Description: openai.String(tool.Description),
					Parameters: openai.FunctionParameters{
						"type":       "object",
						"properties": parameters,
						"required":   required,
					},
				},
			},
		}
	}

	chatCompletionResponse, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Model:    "gpt-4.1",
		Messages: openaiMessages,
		Tools:    openaiTools,
	})
	if err != nil {
		panic(err)
	}
	response := llminterface.ResponseMessageList{}
	if chatCompletionResponse.Choices[0].Message.Content != "" {
		response = append(response, llminterface.AssistantMessage{Content: chatCompletionResponse.Choices[0].Message.Content})
	}
	if chatCompletionResponse.Choices[0].Message.ToolCalls != nil {
		for _, toolCall := range chatCompletionResponse.Choices[0].Message.ToolCalls {
			response = append(response, llminterface.ToolCallMessage{
				ToolCallID: toolCall.ID,
				ToolName:   toolCall.Function.Name,
				Arguments: func() map[string]any {
					var args map[string]any
					json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
					return args
				}(),
			})
		}
	}
	return response
}
