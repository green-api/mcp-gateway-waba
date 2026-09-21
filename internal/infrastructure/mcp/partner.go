package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/green-api/green-api-mcp-gateway-waba/internal/infrastructure/config"
	"github.com/mark3labs/mcp-go/mcp"
	mcpgo "github.com/mark3labs/mcp-go/server"
)

func partnerDo(ctx context.Context, httpMethod, url string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, httpMethod, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("partner API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func registerPartnerTools(s *Server) {
	s.addTool(
		mcp.NewTool(ToolCreateInstance,
			mcp.WithDescription("Create a new instance via the GREEN-API Partner API"),
			mcp.WithString("partner_token", mcp.Required(), mcp.Description("Partner token from your GREEN-API dashboard")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			partnerToken, err := req.RequireString("partner_token")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			url := fmt.Sprintf("%s/partner/createInstance/%s", config.DefaultAPIURL, partnerToken)
			result, err := partnerDo(ctx, http.MethodPost, url, nil)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		}),
	)

	s.addTool(
		mcp.NewTool(ToolDeleteInstance,
			mcp.WithDescription("Delete an instance via the GREEN-API Partner API"),
			mcp.WithString("partner_token", mcp.Required(), mcp.Description("Partner token from your GREEN-API dashboard")),
			instanceArg(),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			partnerToken, err := req.RequireString("partner_token")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			url := fmt.Sprintf("%s/partner/deleteInstance/%s", config.DefaultAPIURL, partnerToken)
			result, err := partnerDo(ctx, http.MethodPost, url, map[string]any{"idInstance": instanceID})
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		}),
	)

	s.addTool(
		mcp.NewTool(ToolGetInstances,
			mcp.WithDescription("List all instances via the GREEN-API Partner API"),
			mcp.WithString("partner_token", mcp.Required(), mcp.Description("Partner token from your GREEN-API dashboard")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			partnerToken, err := req.RequireString("partner_token")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			url := fmt.Sprintf("%s/partner/getInstances/%s", config.DefaultAPIURL, partnerToken)
			result, err := partnerDo(ctx, http.MethodGet, url, nil)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		}),
	)
}
