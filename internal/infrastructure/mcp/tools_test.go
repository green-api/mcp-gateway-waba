package mcp

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/green-api/mcp-gateway-waba/internal/infrastructure"
	"github.com/green-api/mcp-gateway-waba/internal/infrastructure/config"
	mcpgo "github.com/mark3labs/mcp-go/server"
)

func newTestServer() *Server {
	return NewServer(infrastructure.NewCredentialManager(), nil, nil, nil, "test")
}

func TestRegisteredToolsHaveSubmissionReviewHints(t *testing.T) {
	tools := newTestServer().mcp.ListTools()
	if len(tools) == 0 {
		t.Fatal("expected registered tools")
	}

	for name, entry := range tools {
		annotations := entry.Tool.Annotations
		if annotations.ReadOnlyHint == nil {
			t.Errorf("%s missing readOnlyHint", name)
		}
		if annotations.OpenWorldHint == nil {
			t.Errorf("%s missing openWorldHint", name)
		}
		if annotations.DestructiveHint == nil {
			t.Errorf("%s missing destructiveHint", name)
		}
	}

	assertToolHints(t, tools, ToolGetTemplates, true, false, false)
	assertToolHints(t, tools, ToolGetState, true, false, false)
	assertToolHints(t, tools, ToolSendMessage, false, true, true)
	assertToolHints(t, tools, ToolSendTemplate, false, true, true)
	assertToolHints(t, tools, ToolCreateTemplate, false, true, false)
	assertToolHints(t, tools, ToolDeleteTemplate, false, true, true)
	assertToolHints(t, tools, ToolSetSettings, false, true, true)
	assertToolHints(t, tools, ToolDisconnect, false, false, false)
	assertToolHints(t, tools, ToolServiceMethod, false, true, true)
	assertToolHints(t, tools, ToolCreateInstance, false, true, true)
	assertToolHints(t, tools, ToolDeleteInstance, false, true, true)
	assertToolHints(t, tools, ToolGetInstances, true, false, false)
}

func TestRegisteredToolsUsePrefixAndTitles(t *testing.T) {
	tools := newTestServer().mcp.ListTools()
	if len(tools) != 29 {
		t.Errorf("expected 29 WABA tools, got %d", len(tools))
	}
	for name, entry := range tools {
		if !strings.HasPrefix(name, ToolPrefix) {
			t.Errorf("%s does not use the %q prefix", name, ToolPrefix)
		}
		if entry.Tool.Annotations.Title == "" {
			t.Errorf("%s has no title", name)
		}
		if entry.Tool.Description == "" {
			t.Errorf("%s has no description", name)
		}
	}
}

func TestNoWhatsAppOnlyToolsRegistered(t *testing.T) {
	tools := newTestServer().mcp.ListTools()
	for _, name := range []string{"waba_get_qr", "waba_get_contacts", "waba_create_group", "waba_send_poll", "waba_check_whatsapp", "whatsapp_send_message", "waba_call_green_api_method"} {
		if _, ok := tools[name]; ok {
			t.Errorf("%s must not be registered on the WABA gateway", name)
		}
	}
}

func TestServiceMethodToolPointsAtDocsAndHidesItsShape(t *testing.T) {
	tools := newTestServer().mcp.ListTools()
	entry, ok := tools[ToolServiceMethod]
	if !ok {
		t.Fatalf("%s is not registered", ToolServiceMethod)
	}
	if entry.Tool.Description != "https://green-api.com/en/docs/api/" {
		t.Errorf("description = %q, want just the docs link", entry.Tool.Description)
	}

	schema := entry.Tool.InputSchema
	if len(schema.Properties) != 2 {
		t.Fatalf("expected exactly instance_id and request as declared properties, got %v", schema.Properties)
	}
	for _, key := range []string{"instance_id", "request"} {
		if _, ok := schema.Properties[key]; !ok {
			t.Errorf("missing input property %q", key)
		}
	}
	requiredSet := map[string]bool{}
	for _, r := range schema.Required {
		requiredSet[r] = true
	}
	for _, key := range []string{"instance_id", "request"} {
		if !requiredSet[key] {
			t.Errorf("%q should be required", key)
		}
	}

	requestSchema, ok := schema.Properties["request"].(map[string]any)
	if !ok {
		t.Fatalf("request property has unexpected shape: %#v", schema.Properties["request"])
	}
	if props, ok := requestSchema["properties"].(map[string]any); !ok || len(props) != 0 {
		t.Errorf("request must declare no named sub-fields (method/httpMethod/body stay undeclared), got %v", requestSchema["properties"])
	}
	if additional, ok := requestSchema["additionalProperties"].(bool); !ok || !additional {
		t.Errorf("request must allow additional properties, got %v", requestSchema["additionalProperties"])
	}
}

func TestParseServiceRequest(t *testing.T) {
	method, httpMethod, body, err := parseServiceRequest(map[string]any{
		"method":     "sendPoll",
		"httpMethod": "POST",
		"body":       map[string]any{"chatId": "11001234567@c.us"},
	})
	if err != nil {
		t.Fatalf("parseServiceRequest: %v", err)
	}
	if method != "sendPoll" || httpMethod != "POST" {
		t.Errorf("method=%q httpMethod=%q", method, httpMethod)
	}
	bodyMap, ok := body.(map[string]any)
	if !ok || bodyMap["chatId"] != "11001234567@c.us" {
		t.Errorf("body: %v", body)
	}
}

func TestParseServiceRequest_MethodWithoutBody(t *testing.T) {
	method, httpMethod, body, err := parseServiceRequest(map[string]any{"method": "getChats", "httpMethod": "GET"})
	if err != nil {
		t.Fatalf("parseServiceRequest: %v", err)
	}
	if method != "getChats" || httpMethod != "GET" || body != nil {
		t.Errorf("method=%q httpMethod=%q body=%v", method, httpMethod, body)
	}
}

func TestParseServiceRequest_Errors(t *testing.T) {
	if _, _, _, err := parseServiceRequest("not an object"); err == nil {
		t.Error("expected error for a non-object request")
	}
	if _, _, _, err := parseServiceRequest(map[string]any{}); err == nil {
		t.Error("expected error for a missing method")
	}
	if _, _, _, err := parseServiceRequest(nil); err == nil {
		t.Error("expected error for a nil request")
	}
	if _, _, _, err := parseServiceRequest(map[string]any{"method": "getChats"}); err == nil {
		t.Error("expected error for a missing httpMethod")
	}
}

func TestWidgetToolsExposeResourceTemplateMeta(t *testing.T) {
	tools := newTestServer().mcp.ListTools()
	assertToolResourceURI(t, tools, ToolGetTemplates, ResourceTemplatesWidget)

	if entry, ok := tools[ToolSendMessage]; ok && entry.Tool.Meta != nil {
		if _, has := entry.Tool.Meta.AdditionalFields["ui"]; has {
			t.Errorf("%s must not carry widget meta", ToolSendMessage)
		}
	}
}

func assertToolHints(t *testing.T, tools map[string]*mcpgo.ServerTool, name string, readOnly, openWorld, destructive bool) {
	t.Helper()

	entry, ok := tools[name]
	if !ok {
		t.Fatalf("tool %s is not registered", name)
	}
	annotations := entry.Tool.Annotations
	if annotations.ReadOnlyHint == nil || *annotations.ReadOnlyHint != readOnly {
		t.Errorf("%s readOnlyHint = %s, want %v", name, boolPtrString(annotations.ReadOnlyHint), readOnly)
	}
	if annotations.OpenWorldHint == nil || *annotations.OpenWorldHint != openWorld {
		t.Errorf("%s openWorldHint = %s, want %v", name, boolPtrString(annotations.OpenWorldHint), openWorld)
	}
	if annotations.DestructiveHint == nil || *annotations.DestructiveHint != destructive {
		t.Errorf("%s destructiveHint = %s, want %v", name, boolPtrString(annotations.DestructiveHint), destructive)
	}
}

func boolPtrString(value *bool) string {
	if value == nil {
		return "<nil>"
	}
	if *value {
		return "true"
	}
	return "false"
}

func assertToolResourceURI(t *testing.T, tools map[string]*mcpgo.ServerTool, name, want string) {
	t.Helper()

	entry, ok := tools[name]
	if !ok {
		t.Fatalf("tool %s is not registered", name)
	}
	if entry.Tool.Meta == nil {
		t.Fatalf("%s missing _meta", name)
	}
	ui, ok := entry.Tool.Meta.AdditionalFields["ui"].(map[string]any)
	if !ok {
		t.Fatalf("%s missing _meta.ui", name)
	}
	if got := ui["resourceUri"]; got != want {
		t.Fatalf("%s _meta.ui.resourceUri = %v, want %q", name, got, want)
	}
	if got := entry.Tool.Meta.AdditionalFields["openai/outputTemplate"]; got != want {
		t.Fatalf("%s _meta[openai/outputTemplate] = %v, want %q", name, got, want)
	}
}

func TestWidgetResourceMetaHasSubmissionCSPAndDomain(t *testing.T) {
	t.Setenv("GREEN_API_WIDGET_DOMAIN", "https://widgets.green-api.example")

	meta := widgetResourceMeta("Templates widget", []string{"https://cdn.example.com"}, "")

	ui, ok := meta["ui"].(map[string]any)
	if !ok {
		t.Fatalf("ui metadata missing or wrong type: %#v", meta["ui"])
	}
	sum := sha256.Sum256([]byte("https://widgets.green-api.example/mcp"))
	wantDomain := hex.EncodeToString(sum[:])[:32] + ".claudemcpcontent.com"
	if got := ui["domain"]; got != wantDomain {
		t.Fatalf("ui.domain = %v, want %q", got, wantDomain)
	}

	csp, ok := ui["csp"].(map[string]any)
	if !ok {
		t.Fatalf("ui.csp missing or wrong type: %#v", ui["csp"])
	}
	assertStringSlice(t, csp["connectDomains"], []string{config.DefaultAPIURL})
	assertStringSlice(t, csp["resourceDomains"], []string{"https://cdn.example.com"})

	legacyCSP, ok := meta["openai/widgetCSP"].(map[string]any)
	if !ok {
		t.Fatalf("openai/widgetCSP missing or wrong type: %#v", meta["openai/widgetCSP"])
	}
	assertStringSlice(t, legacyCSP["connect_domains"], []string{config.DefaultAPIURL})
	assertStringSlice(t, legacyCSP["resource_domains"], []string{"https://cdn.example.com"})

	if got := meta["openai/widgetDomain"]; got != "https://widgets.green-api.example" {
		t.Fatalf("openai/widgetDomain = %v, want %q", got, "https://widgets.green-api.example")
	}
	if got := meta["openai/widgetDescription"]; got != "Templates widget" {
		t.Fatalf("openai/widgetDescription = %v, want %q", got, "Templates widget")
	}
}

func assertStringSlice(t *testing.T, got any, want []string) {
	t.Helper()

	slice, ok := got.([]string)
	if !ok {
		t.Fatalf("value = %#v, want []string", got)
	}
	if len(slice) != len(want) {
		t.Fatalf("len = %d, want %d for %#v", len(slice), len(want), slice)
	}
	for i := range want {
		if slice[i] != want[i] {
			t.Fatalf("slice[%d] = %q, want %q", i, slice[i], want[i])
		}
	}
}

func TestWidgetResourceMetaClaudeHashDomain(t *testing.T) {
	t.Setenv("GREEN_API_WIDGET_DOMAIN", "abc123.claudemcpcontent.com")

	meta := widgetResourceMeta("Templates widget", nil, "")
	ui := meta["ui"].(map[string]any)
	if got := ui["domain"]; got != "abc123.claudemcpcontent.com" {
		t.Fatalf("ui.domain = %v, want claudemcpcontent hash domain", got)
	}
}

func TestWidgetResourceMetaFallsBackToDefaultDomain(t *testing.T) {
	t.Setenv("GREEN_API_WIDGET_DOMAIN", "")
	t.Setenv("GREEN_API_BASE_URL", "")

	meta := widgetResourceMeta("Templates widget", nil, "")
	if got := meta["openai/widgetDomain"]; got != config.DefaultWidgetDomain {
		t.Fatalf("openai/widgetDomain = %v, want %q", got, config.DefaultWidgetDomain)
	}
}
