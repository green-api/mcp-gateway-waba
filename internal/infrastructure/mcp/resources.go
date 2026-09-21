package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/green-api/green-api-mcp-gateway-waba/internal/domain"
	"github.com/green-api/green-api-mcp-gateway-waba/internal/infrastructure"
	"github.com/green-api/green-api-mcp-gateway-waba/internal/infrastructure/config"
	"github.com/mark3labs/mcp-go/mcp"
	mcpgo "github.com/mark3labs/mcp-go/server"
)

func registerResources(s *Server) {
	templatesResource := mcp.NewResource(ResourceTemplatesWidget, "Templates Widget",
		mcp.WithMIMEType(widgetMIMEType),
		mcp.WithResourceDescription(templatesWidgetDescription),
	)
	templatesResource.Meta = mcp.NewMetaFromMap(widgetResourceMeta(templatesWidgetDescription, []string{}, ""))
	s.mcp.AddResource(
		templatesResource,
		mcpgo.ResourceHandlerFunc(func(ctx context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			meta := widgetResourceMeta(templatesWidgetDescription, []string{}, publicBaseURLFromContext(ctx))
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					Meta:     meta,
					URI:      ResourceTemplatesWidget,
					MIMEType: widgetMIMEType,
					Text:     infrastructure.TemplatesAppHTML,
				},
			}, nil
		}),
	)

	s.addInstanceResource(ResourceInstanceState, "Instance State",
		"Current WABA account state (authorized, notAuthorized, blocked, sleepMode, starting, yellowCard)",
		domain.MethodGetStateInstance, ToolGetState)
	s.addInstanceResource(ResourceInstanceSettings, "Instance Settings",
		"WABA notification settings: webhook URL, token and enabled webhook types",
		domain.MethodGetSettings, ToolGetSettings)
	s.addInstanceResource(ResourceInstanceTemplates, "Instance Templates",
		"All message templates of the WABA account with their Meta review status",
		domain.MethodGetTemplates, ToolGetTemplates)
}

func (s *Server) addInstanceResource(uri, name, description, method, toolName string) {
	s.mcp.AddResourceTemplate(
		mcp.NewResourceTemplate(uri, name,
			mcp.WithTemplateDescription(description),
			mcp.WithTemplateMIMEType("application/json"),
		),
		mcpgo.ResourceTemplateHandlerFunc(func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			instanceID, err := extractInstanceID(req)
			if err != nil {
				return nil, err
			}
			raw, err := s.client.GetMethod(ctx, instanceID, method)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", toolName, err)
			}
			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(raw),
				},
			}, nil
		}),
	)
}

func widgetResourceMeta(description string, resourceDomains []string, requestBaseURL string) map[string]any {
	domainURL := strings.TrimRight(os.Getenv("GREEN_API_WIDGET_DOMAIN"), "/")
	if domainURL == "" {
		domainURL = strings.TrimRight(os.Getenv("GREEN_API_BASE_URL"), "/")
	}
	if domainURL == "" {
		domainURL = strings.TrimRight(requestBaseURL, "/")
	}
	if domainURL == "" {
		domainURL = config.DefaultWidgetDomain
	}

	apiURL := strings.TrimRight(os.Getenv("GREEN_API_URL"), "/")
	if apiURL == "" {
		apiURL = config.DefaultAPIURL
	}
	connectDomains := []string{apiURL}
	standardCSP := map[string]any{
		"connectDomains":  connectDomains,
		"resourceDomains": resourceDomains,
	}
	legacyCSP := map[string]any{
		"connect_domains":  connectDomains,
		"resource_domains": resourceDomains,
	}

	ui := map[string]any{
		"prefersBorder": true,
		"csp":           standardCSP,
		"domain":        claudeWidgetDomain(domainURL),
	}

	return map[string]any{
		"ui":                         ui,
		"openai/widgetDescription":   description,
		"openai/widgetPrefersBorder": true,
		"openai/widgetCSP":           legacyCSP,
		"openai/widgetDomain":        domainURL,
	}
}

func claudeWidgetDomain(baseURL string) string {
	if d := os.Getenv("GREEN_API_WIDGET_DOMAIN"); strings.HasSuffix(d, ".claudemcpcontent.com") {
		return d
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/mcp"
	sum := sha256.Sum256([]byte(endpoint))
	return hex.EncodeToString(sum[:])[:32] + ".claudemcpcontent.com"
}

func extractInstanceID(req mcp.ReadResourceRequest) (uint64, error) {
	args := req.Params.Arguments
	if args == nil {
		return 0, fmt.Errorf("missing URI template arguments (expected {id})")
	}

	raw, ok := args["id"]
	if !ok {
		return 0, fmt.Errorf("missing {id} in resource URI: %s", req.Params.URI)
	}

	switch v := raw.(type) {
	case string:
		var id uint64
		_, err := fmt.Sscanf(v, "%d", &id)
		if err != nil {
			return 0, fmt.Errorf("invalid instance id %q: %w", v, err)
		}
		return id, nil
	case float64:
		return uint64(v), nil
	default:
		return 0, fmt.Errorf("unexpected type for instance id: %T", raw)
	}
}
