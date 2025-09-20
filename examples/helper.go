package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"agentlauncher/internal/llminterface"
	"agentlauncher/launcher"
)

func RegisterTools(agentLauncher *launcher.AgentLauncher) {
	agentLauncher.WithTool("calculate",
		"Calculate the result of the expression a * b + c.",
		func(ctx context.Context, a int, b int, c int) (string, error) {
			return fmt.Sprintf("%d", a*b+c), nil
		},
		[]llminterface.ToolParamSchema{
			{
				Name:        "a",
				Type:        "integer",
				Description: "The first integer.",
				Required:    true,
			},
			{
				Name:        "b",
				Type:        "integer",
				Description: "The second integer.",
				Required:    true,
			},
			{
				Name:        "c",
				Type:        "integer",
				Description: "The third integer.",
				Required:    true,
			},
		})

	agentLauncher.WithTool("get_weather",
		"Get the current weather for a given location.",
		func(ctx context.Context, location string) (string, error) {
			return fmt.Sprintf("The weather in %s is sunny with a high of 75°F.", location), nil
		},
		[]llminterface.ToolParamSchema{
			{
				Name:        "location",
				Type:        "string",
				Description: "The location to get the weather for.",
				Required:    true,
			},
		})

	agentLauncher.WithTool("convert_temperature",
		"Convert temperature from Fahrenheit to Celsius.",
		func(ctx context.Context, fahrenheit float64) (string, error) {
			celsius := (fahrenheit - 32) * 5.0 / 9.0
			return fmt.Sprintf("%.0f°F is %.2f°C.", fahrenheit, celsius), nil
		},
		[]llminterface.ToolParamSchema{
			{
				Name:        "fahrenheit",
				Type:        "number",
				Description: "Temperature in Fahrenheit.",
				Required:    true,
			},
		})

	agentLauncher.WithTool("get_current_time",
		"Get the current date and time.",
		func(ctx context.Context) (string, error) {
			return time.Now().Format("2006-01-02 15:04:05"), nil
		},
		[]llminterface.ToolParamSchema{})

	agentLauncher.WithTool("generate_random_number",
		"Generate a random integer between min and max.",
		func(ctx context.Context, min int, max int) (string, error) {
			if min > max {
				return "", fmt.Errorf("min must be less than or equal to max")
			}
			result := rand.Intn(max-min+1) + min
			return fmt.Sprintf("%d", result), nil
		},
		[]llminterface.ToolParamSchema{
			{
				Name:        "min",
				Type:        "integer",
				Description: "Minimum value.",
				Required:    true,
			},
			{
				Name:        "max",
				Type:        "integer",
				Description: "Maximum value.",
				Required:    true,
			},
		})

	agentLauncher.WithTool("search_web",
		"Search the web for a given query.",
		func(ctx context.Context, query string) (string, error) {
			// Simulate async operation with sleep
			time.Sleep(1 * time.Second)
			return fmt.Sprintf("Search results for '%s': Example result 1, Example result 2.", query), nil
		},
		[]llminterface.ToolParamSchema{
			{
				Name:        "query",
				Type:        "string",
				Description: "The search query.",
				Required:    true,
			},
		})

	agentLauncher.WithTool("get_stock_price",
		"Get the current stock price for a given ticker symbol.",
		func(ctx context.Context, ticker string) (string, error) {
			// Simulate async operation with sleep
			time.Sleep(1 * time.Second)
			return fmt.Sprintf("The current price of %s is $150.00.", ticker), nil
		},
		[]llminterface.ToolParamSchema{
			{
				Name:        "ticker",
				Type:        "string",
				Description: "The stock ticker symbol.",
				Required:    true,
			},
		})

	agentLauncher.WithTool("text_analysis",
		"Analyze the text and provide word count.",
		func(ctx context.Context, text string) (string, error) {
			words := strings.Fields(text)
			wordCount := len(words)
			return fmt.Sprintf("The text contains %d words.", wordCount), nil
		},
		[]llminterface.ToolParamSchema{
			{
				Name:        "text",
				Type:        "string",
				Description: "The text to analyze.",
				Required:    true,
			},
		})

	agentLauncher.WithTool("find_dates",
		"Suggest three suitable dates in the given month (format: YYYY-MM).",
		func(ctx context.Context, month string) (string, error) {
			return fmt.Sprintf("Suggested dates: %s-10, %s-17, %s-24.", month, month, month), nil
		},
		[]llminterface.ToolParamSchema{
			{
				Name:        "month",
				Type:        "string",
				Description: "Month in YYYY-MM format.",
				Required:    true,
			},
		})

	agentLauncher.WithTool("suggest_speakers",
		"Suggest two keynote speakers for a given topic.",
		func(ctx context.Context, topic string) (string, error) {
			return fmt.Sprintf("Keynote speakers in %s: Dr. Alice Smith, Prof. Bob Lee.", topic), nil
		},
		[]llminterface.ToolParamSchema{
			{
				Name:        "topic",
				Type:        "string",
				Description: "The topic for keynote speakers.",
				Required:    true,
			},
		})

	agentLauncher.WithTool("draft_agenda",
		"Prepare a draft agenda with a given number of sessions.",
		func(ctx context.Context, sessions int) (string, error) {
			var agenda strings.Builder
			agenda.WriteString("Draft agenda:\n")
			for i := 0; i < sessions; i++ {
				agenda.WriteString(fmt.Sprintf("Session %d: Topic TBD", i+1))
				if i < sessions-1 {
					agenda.WriteString("\n")
				}
			}
			return agenda.String(), nil
		},
		[]llminterface.ToolParamSchema{
			{
				Name:        "sessions",
				Type:        "integer",
				Description: "Number of sessions.",
				Required:    true,
			},
		})

	agentLauncher.WithTool("list_platforms",
		"List three online platforms suitable for hosting a conference.",
		func(ctx context.Context) (string, error) {
			return "Online platforms: Zoom, Microsoft Teams, Hopin.", nil
		},
		[]llminterface.ToolParamSchema{})

	agentLauncher.WithTool("estimate_budget",
		"Estimate a budget for the event, including speaker fees, platform costs, and marketing.",
		func(ctx context.Context, speakers int, platform string, marketing int) (string, error) {
			speakerFees := speakers * 1000
			platformCost := 500
			total := speakerFees + platformCost + marketing
			return fmt.Sprintf("Estimated budget: Speaker fees $%d, Platform (%s) $%d, Marketing $%d, Total $%d.",
				speakerFees, platform, platformCost, marketing, total), nil
		},
		[]llminterface.ToolParamSchema{
			{
				Name:        "speakers",
				Type:        "integer",
				Description: "Number of speakers.",
				Required:    true,
			},
			{
				Name:        "platform",
				Type:        "string",
				Description: "Platform name.",
				Required:    true,
			},
			{
				Name:        "marketing",
				Type:        "integer",
				Description: "Marketing budget in USD.",
				Required:    true,
			},
		})

	agentLauncher.WithTool("draft_email",
		"Draft an invitation email for the event.",
		func(ctx context.Context, eventName string) (string, error) {
			return fmt.Sprintf("Subject: Invitation to %s\nDear Attendee,\nYou are invited to our virtual conference. More details to follow.",
				eventName), nil
		},
		[]llminterface.ToolParamSchema{
			{
				Name:        "event_name",
				Type:        "string",
				Description: "Name of the event.",
				Required:    true,
			},
		})
}

func RegisterMessageHandlers(agentLauncher *launcher.AgentLauncher) {
	agentLauncher.WithConversationHandler(func(messages llminterface.MessageList) llminterface.MessageList {
		fmt.Println("Response Messages: ", len(messages))
		return messages
	})
}
