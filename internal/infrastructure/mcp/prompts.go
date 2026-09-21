package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpgo "github.com/mark3labs/mcp-go/server"
)

func registerPrompts(s *Server) {
	s.mcp.AddPrompt(
		mcp.NewPrompt(PromptCustomerSupport,
			mcp.WithPromptDescription("System prompt for a customer support agent on an official WhatsApp Business (WABA) number. Covers the 24-hour window and the template fallback."),
			mcp.WithArgument("instance_id",
				mcp.RequiredArgument(),
				mcp.ArgumentDescription("WABA instance ID used for support"),
			),
			mcp.WithArgument("language",
				mcp.RequiredArgument(),
				mcp.ArgumentDescription("Language to use with customers (e.g. en, ru, es)"),
			),
		),
		mcpgo.PromptHandlerFunc(func(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			instanceID := req.Params.Arguments["instance_id"]
			language := req.Params.Arguments["language"]
			if instanceID == "" {
				return nil, fmt.Errorf("instance_id is required")
			}
			if language == "" {
				return nil, fmt.Errorf("language is required")
			}

			text := fmt.Sprintf(customerSupportPrompt,
				instanceID, language,
				ToolLastIncomingMessages, ToolSendMessage, ToolSendFile,
				ToolGetTemplates, ToolSendTemplate, ToolGetChatHistory)

			return &mcp.GetPromptResult{
				Description: fmt.Sprintf("WABA customer support prompt for instance %s (%s)", instanceID, language),
				Messages: []mcp.PromptMessage{
					{
						Role:    mcp.RoleUser,
						Content: mcp.TextContent{Type: "text", Text: text},
					},
				},
			}, nil
		}),
	)

	s.mcp.AddPrompt(
		mcp.NewPrompt(PromptTemplateBroadcast,
			mcp.WithPromptDescription("Plan and run a template broadcast to a list of opted-in recipients with an APPROVED WABA template."),
			mcp.WithArgument("instance_id",
				mcp.RequiredArgument(),
				mcp.ArgumentDescription("WABA instance ID used for the broadcast"),
			),
			mcp.WithArgument("template_id",
				mcp.RequiredArgument(),
				mcp.ArgumentDescription("templateId of the APPROVED template to send"),
			),
		),
		mcpgo.PromptHandlerFunc(func(_ context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			instanceID := req.Params.Arguments["instance_id"]
			templateID := req.Params.Arguments["template_id"]
			if instanceID == "" {
				return nil, fmt.Errorf("instance_id is required")
			}
			if templateID == "" {
				return nil, fmt.Errorf("template_id is required")
			}

			text := fmt.Sprintf(templateBroadcastPrompt, instanceID, templateID, ToolGetTemplate, ToolSendTemplate)

			return &mcp.GetPromptResult{
				Description: fmt.Sprintf("WABA template broadcast prompt for instance %s", instanceID),
				Messages: []mcp.PromptMessage{
					{
						Role:    mcp.RoleUser,
						Content: mcp.TextContent{Type: "text", Text: text},
					},
				},
			}, nil
		}),
	)
}
