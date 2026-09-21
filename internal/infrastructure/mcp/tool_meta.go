package mcp

import "github.com/mark3labs/mcp-go/mcp"

const (
	widgetMIMEType             = "text/html;profile=mcp-app"
	templatesWidgetDescription = "Interactive list of WABA message templates with status filter, template details and quick actions"

	instanceIDHint = "WABA instance ID from console.green-api.com"
	chatIDHint     = "Recipient chat ID: phone number in international format without plus or spaces followed by @c.us, for example 11001234567@c.us. Group chats are not supported by WABA"
	windowHint     = "Meta delivers free-form messages only inside the 24-hour customer service window opened by the customer's last incoming message. Outside that window use " + ToolSendTemplate + " with an APPROVED template"

	greenAPIDocsURL   = "https://green-api.com/en/docs/api/"
	serviceMethodHint = greenAPIDocsURL
)

var toolTitles = map[string]string{
	ToolConnect:              "Connect Instance",
	ToolDisconnect:           "Disconnect Instance",
	ToolSendMessage:          "Send Message",
	ToolSendTemplate:         "Send Template",
	ToolSendFile:             "Send File by URL",
	ToolSendFileByUpload:     "Send File",
	ToolGetState:             "Get Account State",
	ToolGetSettings:          "Get Settings",
	ToolSetSettings:          "Update Settings",
	ToolGetWaSettings:        "Get Account Info",
	ToolReboot:               "Reboot Instance",
	ToolCreateTemplate:       "Create Template",
	ToolEditTemplate:         "Edit Template",
	ToolGetTemplates:         "List Templates",
	ToolGetTemplate:          "Get Template",
	ToolDeleteTemplate:       "Delete Template",
	ToolDeleteTemplateByID:   "Delete Template by ID",
	ToolReceiveNotification:  "Receive Notification",
	ToolDeleteNotification:   "Delete Notification",
	ToolGetChatHistory:       "Get Chat History",
	ToolGetMessage:           "Get Message",
	ToolLastIncomingMessages: "Get Recent Incoming Messages",
	ToolLastOutgoingMessages: "Get Recent Outgoing Messages",
	ToolShowMessagesQueue:    "Show Messages Queue",
	ToolClearMessagesQueue:   "Clear Messages Queue",
	ToolServiceMethod:        "Service Method",
	ToolCreateInstance:       "Create Partner Instance",
	ToolDeleteInstance:       "Delete Partner Instance",
	ToolGetInstances:         "List Partner Instances",
}

var readOnlyTools = map[string]bool{
	ToolGetState:             true,
	ToolGetSettings:          true,
	ToolGetWaSettings:        true,
	ToolGetTemplates:         true,
	ToolGetTemplate:          true,
	ToolReceiveNotification:  true,
	ToolGetChatHistory:       true,
	ToolGetMessage:           true,
	ToolLastIncomingMessages: true,
	ToolLastOutgoingMessages: true,
	ToolShowMessagesQueue:    true,
	ToolGetInstances:         true,
}

var localOnlyTools = map[string]bool{
	ToolConnect:    true,
	ToolDisconnect: true,
}

var destructiveTools = map[string]bool{
	ToolSendMessage:        true,
	ToolSendTemplate:       true,
	ToolSendFile:           true,
	ToolSendFileByUpload:   true,
	ToolSetSettings:        true,
	ToolReboot:             true,
	ToolEditTemplate:       true,
	ToolDeleteTemplate:     true,
	ToolDeleteTemplateByID: true,
	ToolDeleteNotification: true,
	ToolClearMessagesQueue: true,
	ToolServiceMethod:      true,
	ToolCreateInstance:     true,
	ToolDeleteInstance:     true,
}

var widgetResourceByTool = map[string]string{
	ToolGetTemplates: ResourceTemplatesWidget,
}

var templateButtonSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"type":         map[string]any{"type": "string", "enum": []string{"QUICK_REPLY", "URL", "PHONE_NUMBER", "OTP"}},
		"text":         map[string]any{"type": "string"},
		"url":          map[string]any{"type": "string"},
		"phone_number": map[string]any{"type": "string"},
		"otp_type":     map[string]any{"type": "string"},
		"example":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
	},
	"required": []string{"type"},
}

var templateCardSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"headerType": map[string]any{"type": "string"},
		"mediaUrl":   map[string]any{"type": "string"},
		"body":       map[string]any{"type": "string"},
		"sampleText": map[string]any{"type": "string"},
		"buttons":    map[string]any{"type": "array", "items": templateButtonSchema},
	},
	"required": []string{"headerType", "mediaUrl"},
}

var postbackTextSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"index": map[string]any{"type": "integer"},
		"text":  map[string]any{"type": "string"},
	},
	"required": []string{"index", "text"},
}

func applySubmissionReviewHints(tool *mcp.Tool) {
	readOnly := readOnlyTools[tool.Name]
	openWorld := !readOnly && !localOnlyTools[tool.Name]
	destructive := destructiveTools[tool.Name]

	tool.Annotations.ReadOnlyHint = &readOnly
	tool.Annotations.OpenWorldHint = &openWorld
	tool.Annotations.DestructiveHint = &destructive
	if readOnly {
		idempotent := true
		tool.Annotations.IdempotentHint = &idempotent
	}
}

func applyToolTitle(tool *mcp.Tool) {
	if title, ok := toolTitles[tool.Name]; ok && tool.Annotations.Title == "" {
		tool.Annotations.Title = title
	}
}

func applyWidgetToolMeta(tool *mcp.Tool) {
	resourceURI, ok := widgetResourceByTool[tool.Name]
	if !ok {
		return
	}

	fields := map[string]any{}
	if tool.Meta != nil {
		for key, value := range tool.Meta.AdditionalFields {
			fields[key] = value
		}
	}
	fields["ui"] = map[string]any{"resourceUri": resourceURI}
	fields["openai/outputTemplate"] = resourceURI
	tool.Meta = mcp.NewMetaFromMap(fields)
}
