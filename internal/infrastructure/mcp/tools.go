package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/green-api/mcp-gateway-waba/internal/domain"
	"github.com/green-api/mcp-gateway-waba/internal/infrastructure"
	"github.com/green-api/mcp-gateway-waba/internal/infrastructure/config"
	"github.com/mark3labs/mcp-go/mcp"
	mcpgo "github.com/mark3labs/mcp-go/server"
)

func registerTools(s *Server) {
	registerSessionTools(s)
	registerSendingTools(s)
	registerAccountTools(s)
	registerTemplateTools(s)
	registerReceivingTools(s)
	registerJournalTools(s)
	registerQueueTools(s)
	registerServiceMethodTool(s)
	registerPartnerTools(s)
}

func instanceArg() mcp.ToolOption {
	return mcp.WithNumber("instance_id", mcp.Required(), mcp.Description(instanceIDHint))
}

func (s *Server) addGetTool(tool mcp.Tool, method, emptyResult string) {
	s.addTool(tool, mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		instanceID, err := resolveInstanceID(ctx, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		result, err := s.client.GetMethod(ctx, instanceID, method)
		return rawResultOr(emptyResult, result, err)
	}))
}

func registerSessionTools(s *Server) {
	s.addTool(
		mcp.NewTool(ToolConnect,
			mcp.WithDescription("Register a GREEN-API WABA (official WhatsApp Business API) instance for this session and validate its credentials. Not needed on the hosted server, where OAuth binds the instance automatically."),
			instanceArg(),
			mcp.WithString("api_token", mcp.Required(), mcp.Description("Instance API token from console.green-api.com")),
			mcp.WithString("api_url", mcp.Description("API base URL, defaults to "+config.DefaultAPIURL)),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			apiToken, err := req.RequireString("api_token")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			s.credentials.AddInstance(&domain.InstanceCredentials{
				InstanceID: instanceID,
				APIToken:   apiToken,
				APIURL:     req.GetString("api_url", config.DefaultAPIURL),
			})
			state, err := s.client.GetMethod(ctx, instanceID, domain.MethodGetStateInstance)
			if err != nil {
				s.credentials.RemoveInstance(instanceID)
				return mcp.NewToolResultError(fmt.Sprintf("Connection failed: %s", err.Error())), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf(`{"connected": true, "instance_id": %d, "state": %s}`, instanceID, string(state))), nil
		}),
	)

	s.addTool(
		mcp.NewTool(ToolDisconnect,
			mcp.WithDescription("Remove a WABA instance and its credentials from this session"),
			instanceArg(),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			s.credentials.RemoveInstance(instanceID)
			return mcp.NewToolResultText(fmt.Sprintf(`{"disconnected": true, "instance_id": %d}`, instanceID)), nil
		}),
	)
}

func registerSendingTools(s *Server) {
	s.addTool(
		mcp.NewTool(ToolSendMessage,
			mcp.WithDescription("Send a free-form text message of up to 20000 characters. "+windowHint+"."),
			instanceArg(),
			mcp.WithString("chat_id", mcp.Required(), mcp.Description(chatIDHint)),
			mcp.WithString("message", mcp.Required(), mcp.Description("Message text, emoji supported")),
			mcp.WithBoolean("link_preview", mcp.Description("Show a preview for links in the text, enabled by default")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vals, err := requireStrings(req, "chat_id", "message")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := domain.SendMessageRequest{ChatID: vals["chat_id"], Message: vals["message"]}
			if v, ok := req.GetArguments()["link_preview"].(bool); ok {
				body.LinkPreview = &v
			}
			return rawResult(s.client.CallMethod(ctx, instanceID, domain.MethodSendMessage, body))
		}),
	)

	s.addTool(
		mcp.NewTool(ToolSendTemplate,
			mcp.WithDescription("Send a pre-approved message template. This is the only way to start a conversation or to message a customer outside the 24-hour window. Get template_id from "+ToolGetTemplates+"; the template status must be APPROVED. params fill the placeholders {{1}}, {{2}}... in order. For IMAGE, VIDEO or DOCUMENT templates pass media_type and media_url; for LOCATION or CAROUSEL templates pass the raw message object."),
			instanceArg(),
			mcp.WithString("chat_id", mcp.Required(), mcp.Description(chatIDHint)),
			mcp.WithString("template_id", mcp.Required(), mcp.Description("templateId of an APPROVED template")),
			mcp.WithArray("params", mcp.Items(map[string]any{"type": "string"}), mcp.Description("Values for the template placeholders {{1}}, {{2}}... in order")),
			mcp.WithString("media_type", mcp.Enum(domain.TemplateMediaTypes...), mcp.Description("Header media type for IMAGE, VIDEO or DOCUMENT templates")),
			mcp.WithString("media_url", mcp.Description("Public URL of the header media")),
			mcp.WithString("media_filename", mcp.Description("File name with extension, DOCUMENT templates only")),
			mcp.WithObject("message", mcp.AdditionalProperties(true), mcp.Description("Raw message object as in the sendTemplate docs, for LOCATION or CAROUSEL templates. Overrides the media_* arguments")),
			mcp.WithArray("postback_texts", mcp.Items(postbackTextSchema), mcp.Description("Postback payloads for quick-reply buttons, one object per button: {index, text}")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vals, err := requireStrings(req, "chat_id", "template_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := domain.SendTemplateRequest{
				ChatID:     vals["chat_id"],
				TemplateID: vals["template_id"],
				Params:     req.GetStringSlice("params", nil),
			}
			args := req.GetArguments()
			if raw, ok := args["message"]; ok && raw != nil {
				if err := decodeArg(raw, &body.Message); err != nil {
					return mcp.NewToolResultError("message: " + err.Error()), nil
				}
			} else if mediaType := strings.ToLower(req.GetString("media_type", "")); mediaType != "" {
				body.Message = templateMediaMessage(mediaType, req.GetString("media_url", ""), req.GetString("media_filename", ""))
			}
			if raw, ok := args["postback_texts"]; ok && raw != nil {
				if err := decodeArg(raw, &body.PostbackTexts); err != nil {
					return mcp.NewToolResultError("postback_texts: " + err.Error()), nil
				}
			}
			return rawResult(s.client.CallMethod(ctx, instanceID, domain.MethodSendTemplate, body))
		}),
	)

	s.addTool(
		mcp.NewTool(ToolSendFile,
			mcp.WithDescription("Send an image, video, audio file or document by public URL. "+windowHint+". Limits: images 5 MB, audio and video 16 MB, documents 100 MB."),
			instanceArg(),
			mcp.WithString("chat_id", mcp.Required(), mcp.Description(chatIDHint)),
			mcp.WithString("url", mcp.Required(), mcp.Description("Public URL of the file")),
			mcp.WithString("file_name", mcp.Required(), mcp.Description("File name with extension, for example invoice.pdf")),
			mcp.WithString("caption", mcp.Description("Caption for images and videos, up to 20000 characters")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vals, err := requireStrings(req, "chat_id", "url", "file_name")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := domain.SendFileByURLRequest{
				ChatID:   vals["chat_id"],
				URL:      vals["url"],
				FileName: vals["file_name"],
				Caption:  req.GetString("caption", ""),
			}
			return rawResult(s.client.CallMethod(ctx, instanceID, domain.MethodSendFileByURL, body))
		}),
	)

	s.addTool(
		mcp.NewTool(ToolSendFileByUpload,
			mcp.WithDescription("Send a file from base64 contents via multipart upload, no public URL needed. "+windowHint+". Returns idMessage and a temporary urlFile."),
			instanceArg(),
			mcp.WithString("chat_id", mcp.Required(), mcp.Description(chatIDHint)),
			mcp.WithString("file_base64", mcp.Required(), mcp.Description("File contents encoded as base64")),
			mcp.WithString("filename", mcp.Required(), mcp.Description("File name with extension, for example photo.jpg")),
			mcp.WithString("caption", mcp.Description("Caption for images and videos")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vals, err := requireStrings(req, "chat_id", "file_base64", "filename")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			fileData, err := base64.StdEncoding.DecodeString(vals["file_base64"])
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid base64: %s", err.Error())), nil
			}
			resp, err := s.client.SendFileByUpload(ctx, instanceID, vals["chat_id"], req.GetString("caption", ""), vals["filename"], fileData)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(resp)
		}),
	)
}

func registerAccountTools(s *Server) {
	s.addGetTool(
		mcp.NewTool(ToolGetState,
			mcp.WithDescription("Get the WABA account state: authorized, notAuthorized, blocked, sleepMode, starting or yellowCard (sending suspended by Meta for spam). WABA instances are connected in console.green-api.com, there is no QR code flow."),
			instanceArg(),
		),
		domain.MethodGetStateInstance, `{}`,
	)

	s.addGetTool(
		mcp.NewTool(ToolGetSettings,
			mcp.WithDescription("Get notification settings: webhook URL and token, and which webhook types are enabled"),
			instanceArg(),
		),
		domain.MethodGetSettings, `{}`,
	)

	s.addTool(
		mcp.NewTool(ToolSetSettings,
			mcp.WithDescription("Update notification settings. Only these fields exist for WABA; pass an empty webhook_url to disable webhooks. At least one field must be given."),
			instanceArg(),
			mcp.WithString("webhook_url", mcp.Description("URL that receives notifications, empty string disables webhooks")),
			mcp.WithString("webhook_url_token", mcp.Description("Authorization header value sent with each webhook, empty string to clear")),
			mcp.WithBoolean("outgoing_webhook", mcp.Description("Receive status updates for outgoing messages")),
			mcp.WithBoolean("outgoing_api_message_webhook", mcp.Description("Receive notifications for messages sent via the API")),
			mcp.WithBoolean("incoming_webhook", mcp.Description("Receive incoming messages and files")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			body := domain.SetSettingsRequest{}
			if v, ok := args["webhook_url"].(string); ok {
				body.WebhookURL = &v
			}
			if v, ok := args["webhook_url_token"].(string); ok {
				body.WebhookURLToken = &v
			}
			if v, ok := args["outgoing_webhook"].(bool); ok {
				body.OutgoingWebhook = yesNo(v)
			}
			if v, ok := args["outgoing_api_message_webhook"].(bool); ok {
				body.OutgoingAPIMessageWebhook = yesNo(v)
			}
			if v, ok := args["incoming_webhook"].(bool); ok {
				body.IncomingWebhook = yesNo(v)
			}
			if body == (domain.SetSettingsRequest{}) {
				return mcp.NewToolResultError("at least one setting must be provided"), nil
			}
			return rawResult(s.client.CallMethod(ctx, instanceID, domain.MethodSetSettings, body))
		}),
	)

	s.addGetTool(
		mcp.NewTool(ToolGetWaSettings,
			mcp.WithDescription("Get the WhatsApp Business account phone number and state"),
			instanceArg(),
		),
		domain.MethodGetWaSettings, `{}`,
	)

	s.addGetTool(
		mcp.NewTool(ToolReboot,
			mcp.WithDescription("Reboot the WABA instance"),
			instanceArg(),
		),
		domain.MethodReboot, `{"isReboot": true}`,
	)
}

func registerTemplateTools(s *Server) {
	s.addTool(
		mcp.NewTool(ToolCreateTemplate,
			mcp.WithDescription("Create a message template and submit it to Meta for review. The template comes back with status PENDING and can be sent with "+ToolSendTemplate+" only after it becomes APPROVED (check with "+ToolGetTemplates+"). Placeholders in content are {{1}}, {{2}}... and example must show them filled in. MARKETING and AUTHENTICATION templates are billed per message, UTILITY templates are cheaper."),
			instanceArg(),
			mcp.WithString("element_name", mcp.Required(), mcp.Description("Unique template name: lowercase letters, digits and underscores only")),
			mcp.WithString("language_code", mcp.Required(), mcp.Description("Template language code, for example en, en_US, ru, pt_BR")),
			mcp.WithString("category", mcp.Required(), mcp.Enum(domain.TemplateCategories...), mcp.Description("AUTHENTICATION, MARKETING or UTILITY")),
			mcp.WithString("template_type", mcp.Required(), mcp.Enum(domain.TemplateTypes...), mcp.Description("TEXT, IMAGE, VIDEO, DOCUMENT or CAROUSEL")),
			mcp.WithString("vertical", mcp.Required(), mcp.Description("Short description of the template purpose for Meta review, max 180 characters")),
			mcp.WithString("content", mcp.Description("Template body, max 550 characters, placeholders {{1}}, {{2}}. Optional for AUTHENTICATION")),
			mcp.WithString("example", mcp.Description("The body with example values substituted for every placeholder. Required unless category is AUTHENTICATION")),
			mcp.WithString("header", mcp.Description("Header text, max 60 characters, TEXT templates only")),
			mcp.WithString("example_header", mcp.Description("Example header text, required when header is set")),
			mcp.WithString("footer", mcp.Description("Footer text, max 60 characters")),
			mcp.WithArray("buttons", mcp.Items(templateButtonSchema), mcp.Description("Buttons: type QUICK_REPLY, URL, PHONE_NUMBER or OTP with text, url, phone_number, otp_type and example. AUTHENTICATION templates need one OTP button")),
			mcp.WithArray("cards", mcp.Items(templateCardSchema), mcp.Description("CAROUSEL cards, 2 to 10, each with headerType IMAGE, mediaUrl, body, sampleText and up to 2 buttons")),
			mcp.WithString("media_url", mcp.Description("Sample media URL for IMAGE, VIDEO and DOCUMENT templates")),
			mcp.WithBoolean("enable_sample", mcp.DefaultBool(true), mcp.Description("Submit the example values as the Meta sample, keep true")),
			mcp.WithBoolean("add_security_recommendation", mcp.Description("AUTHENTICATION only: add the security disclaimer to the body")),
			mcp.WithNumber("code_expiration_minutes", mcp.Description("AUTHENTICATION only: code expiry note in the footer, 1 to 90")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vals, err := requireStrings(req, "element_name", "language_code", "category", "template_type", "vertical")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			buttons, err := jsonStringArg(args, "buttons")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			cards, err := jsonStringArg(args, "cards")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := domain.CreateTemplateRequest{
				ElementName:               vals["element_name"],
				LanguageCode:              vals["language_code"],
				Category:                  strings.ToUpper(vals["category"]),
				TemplateType:              strings.ToUpper(vals["template_type"]),
				Vertical:                  vals["vertical"],
				Content:                   req.GetString("content", ""),
				Example:                   req.GetString("example", ""),
				Header:                    req.GetString("header", ""),
				ExampleHeader:             req.GetString("example_header", ""),
				Footer:                    req.GetString("footer", ""),
				Buttons:                   buttons,
				Cards:                     cards,
				MediaURL:                  req.GetString("media_url", ""),
				EnableSample:              req.GetBool("enable_sample", true),
				AddSecurityRecommendation: req.GetBool("add_security_recommendation", false),
				CodeExpirationMinutes:     req.GetInt("code_expiration_minutes", 0),
			}
			return rawResult(s.client.CallMethod(ctx, instanceID, domain.MethodCreateTemplate, body))
		}),
	)

	s.addTool(
		mcp.NewTool(ToolEditTemplate,
			mcp.WithDescription("Edit an existing template. Only templates in REJECTED, APPROVED or PAUSED status can be edited; after editing the template returns to PENDING until Meta reviews it again. Pass only the fields to change."),
			instanceArg(),
			mcp.WithString("template_id", mcp.Required(), mcp.Description("templateId from "+ToolGetTemplates)),
			mcp.WithString("content", mcp.Description("New body, max 550 characters, placeholders {{1}}, {{2}}")),
			mcp.WithString("example", mcp.Description("The new body with example values substituted")),
			mcp.WithString("template_type", mcp.Enum(domain.TemplateTypeText, domain.TemplateTypeImage, domain.TemplateTypeVideo, domain.TemplateTypeDocument), mcp.Description("TEXT, IMAGE, VIDEO or DOCUMENT")),
			mcp.WithString("header", mcp.Description("Header text, max 60 characters, TEXT templates only")),
			mcp.WithString("example_header", mcp.Description("Example header text, required when header is set")),
			mcp.WithString("footer", mcp.Description("Footer text, max 60 characters")),
			mcp.WithArray("buttons", mcp.Items(templateButtonSchema), mcp.Description("Replacement buttons: type QUICK_REPLY, URL or PHONE_NUMBER with text, url or phone_number")),
			mcp.WithString("media_url", mcp.Description("Sample media URL for IMAGE, VIDEO and DOCUMENT templates")),
			mcp.WithString("category", mcp.Enum(domain.TemplateCategories...), mcp.Description("AUTHENTICATION, MARKETING or UTILITY")),
			mcp.WithBoolean("enable_sample", mcp.Description("Submit the example values as the Meta sample")),
			mcp.WithBoolean("allow_template_category_change", mcp.Description("Let Meta change the category during review")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			templateID, err := req.RequireString("template_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			args := req.GetArguments()
			buttons, err := jsonStringArg(args, "buttons")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			params := domain.EditTemplateParams{
				Content:       req.GetString("content", ""),
				TemplateType:  strings.ToUpper(req.GetString("template_type", "")),
				Example:       req.GetString("example", ""),
				Header:        req.GetString("header", ""),
				ExampleHeader: req.GetString("example_header", ""),
				Footer:        req.GetString("footer", ""),
				Buttons:       buttons,
				MediaURL:      req.GetString("media_url", ""),
				Category:      strings.ToUpper(req.GetString("category", "")),
			}
			if v, ok := args["enable_sample"].(bool); ok {
				params.EnableSample = &v
			}
			if v, ok := args["allow_template_category_change"].(bool); ok {
				params.AllowTemplateCategoryChange = &v
			}
			body := domain.EditTemplateRequest{TemplateID: templateID, TemplateParams: params}
			return rawResult(s.client.CallMethod(ctx, instanceID, domain.MethodEditTemplate, body))
		}),
	)

	s.addTool(
		mcp.NewTool(ToolGetTemplates,
			mcp.WithDescription("List all message templates of the account with their Meta review status (PENDING, APPROVED, REJECTED, FAILED, PAUSED). Only APPROVED templates can be sent. Optionally filter by status."),
			instanceArg(),
			mcp.WithString("status", mcp.Enum(domain.TemplateStatuses...), mcp.Description("Return only templates with this status")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			status := strings.ToUpper(req.GetString("status", ""))
			if status == "" {
				return rawResult(s.client.GetMethod(ctx, instanceID, domain.MethodGetTemplates))
			}
			resp, err := s.client.GetTemplates(ctx, instanceID)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			filtered := make([]domain.Template, 0, len(resp.Templates))
			for _, tpl := range resp.Templates {
				if tpl.Status == status {
					filtered = append(filtered, tpl)
				}
			}
			return jsonResult(domain.GetTemplatesResponse{Templates: filtered})
		}),
	)

	s.addTool(
		mcp.NewTool(ToolGetTemplate,
			mcp.WithDescription("Get one template by its templateId, including status, body and buttons"),
			instanceArg(),
			mcp.WithString("template_id", mcp.Required(), mcp.Description("templateId from "+ToolGetTemplates)),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			templateID, err := req.RequireString("template_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := domain.GetTemplateByIDRequest{TemplateID: templateID}
			return rawResult(s.client.CallMethod(ctx, instanceID, domain.MethodGetTemplateByID, body))
		}),
	)

	s.addTool(
		mcp.NewTool(ToolDeleteTemplate,
			mcp.WithDescription("Delete all language versions of a template by its element name"),
			instanceArg(),
			mcp.WithString("element_name", mcp.Required(), mcp.Description("elementName of the template, see "+ToolGetTemplates)),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			elementName, err := req.RequireString("element_name")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := domain.DeleteTemplateRequest{ElementName: elementName}
			return rawResult(s.client.CallMethod(ctx, instanceID, domain.MethodDeleteTemplate, body))
		}),
	)

	s.addTool(
		mcp.NewTool(ToolDeleteTemplateByID,
			mcp.WithDescription("Delete one template version by element name and templateId"),
			instanceArg(),
			mcp.WithString("element_name", mcp.Required(), mcp.Description("elementName of the template")),
			mcp.WithString("template_id", mcp.Required(), mcp.Description("templateId of the version to delete")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vals, err := requireStrings(req, "element_name", "template_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := domain.DeleteTemplateByIDRequest{ElementName: vals["element_name"], TemplateID: vals["template_id"]}
			return rawResult(s.client.CallMethod(ctx, instanceID, domain.MethodDeleteTemplateByID, body))
		}),
	)
}

func registerReceivingTools(s *Server) {
	s.addGetTool(
		mcp.NewTool(ToolReceiveNotification,
			mcp.WithDescription("Take one pending notification from the queue: incoming messages, template button replies, delivery statuses. Notifications wait up to 24 hours. Acknowledge each one with "+ToolDeleteNotification+" using its receiptId, otherwise it is returned again."),
			instanceArg(),
		),
		domain.MethodReceiveNotification, `{"notification": null}`,
	)

	s.addTool(
		mcp.NewTool(ToolDeleteNotification,
			mcp.WithDescription("Acknowledge and remove a notification from the queue"),
			instanceArg(),
			mcp.WithNumber("receipt_id", mcp.Required(), mcp.Description("receiptId from "+ToolReceiveNotification)),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			receiptID, err := req.RequireFloat("receipt_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if err := s.client.DeleteNotification(ctx, instanceID, int(receiptID)); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(`{"deleted": true}`), nil
		}),
	)
}

func registerJournalTools(s *Server) {
	s.addTool(
		mcp.NewTool(ToolGetChatHistory,
			mcp.WithDescription("Get the message history of a chat, newest first. Keep count small (default 50, max 100)."),
			instanceArg(),
			mcp.WithString("chat_id", mcp.Required(), mcp.Description(chatIDHint)),
			mcp.WithNumber("count", mcp.Description("Number of messages, default 50, max 100")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			chatID, err := req.RequireString("chat_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := domain.GetChatHistoryRequest{ChatID: chatID, Count: clampInt(req.GetInt("count", 0), 50, 100)}
			return rawResult(s.client.GetChatHistory(ctx, instanceID, body))
		}),
	)

	s.addTool(
		mcp.NewTool(ToolGetMessage,
			mcp.WithDescription("Get one message of a chat by its ID"),
			instanceArg(),
			mcp.WithString("chat_id", mcp.Required(), mcp.Description(chatIDHint)),
			mcp.WithString("id_message", mcp.Required(), mcp.Description("Message ID")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			vals, err := requireStrings(req, "chat_id", "id_message")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := domain.GetMessageRequest{ChatID: vals["chat_id"], IDMessage: vals["id_message"]}
			return rawResult(s.client.GetMessage(ctx, instanceID, body))
		}),
	)

	s.addTool(
		mcp.NewTool(ToolLastIncomingMessages,
			mcp.WithDescription("Get incoming messages from the last N minutes (default 60, max 1440). Useful to see which customers opened a 24-hour window."),
			instanceArg(),
			mcp.WithNumber("minutes", mcp.Description("Time window in minutes, default 60, max 1440")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			minutes := clampInt(req.GetInt("minutes", 0), 60, 1440)
			result, err := s.client.LastIncomingMessages(ctx, instanceID, minutes)
			return rawResultOr(`[]`, result, err)
		}),
	)

	s.addTool(
		mcp.NewTool(ToolLastOutgoingMessages,
			mcp.WithDescription("Get outgoing messages from the last N minutes (default 60, max 1440) with their delivery status"),
			instanceArg(),
			mcp.WithNumber("minutes", mcp.Description("Time window in minutes, default 60, max 1440")),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			minutes := clampInt(req.GetInt("minutes", 0), 60, 1440)
			result, err := s.client.LastOutgoingMessages(ctx, instanceID, minutes)
			return rawResultOr(`[]`, result, err)
		}),
	)
}

func registerQueueTools(s *Server) {
	s.addGetTool(
		mcp.NewTool(ToolShowMessagesQueue,
			mcp.WithDescription("List messages waiting in the outgoing send queue"),
			instanceArg(),
		),
		domain.MethodShowMessagesQueue, `[]`,
	)

	s.addGetTool(
		mcp.NewTool(ToolClearMessagesQueue,
			mcp.WithDescription("Drop every message waiting in the outgoing send queue"),
			instanceArg(),
		),
		domain.MethodClearMessagesQueue, `{"isCleared": true}`,
	)
}

func registerServiceMethodTool(s *Server) {
	s.addTool(
		mcp.NewTool(ToolServiceMethod,
			mcp.WithDescription(serviceMethodHint),
			instanceArg(),
			mcp.WithObject("request", mcp.Required(), mcp.AdditionalProperties(true)),
		),
		mcpgo.ToolHandlerFunc(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			instanceID, err := resolveInstanceID(ctx, req)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			method, httpMethod, body, err := parseServiceRequest(req.GetArguments()["request"])
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return rawResult(s.client.RawMethod(ctx, instanceID, httpMethod, method, body))
		}),
	)
}

// The {method, httpMethod, body} shape is deliberately undeclared in the tool's
// JSON schema; don't add it there. A caller learns it from the error text below.
func parseServiceRequest(raw any) (method, httpMethod string, body any, err error) {
	obj, ok := raw.(map[string]any)
	if !ok {
		return "", "", nil, fmt.Errorf(`request must be an object shaped like {"method": "sendPoll", "httpMethod": "POST", "body": {...}}; see %s for the method name, verb and request body`, greenAPIDocsURL)
	}
	method, _ = obj["method"].(string)
	if method == "" {
		return "", "", nil, fmt.Errorf(`request.method is required, e.g. "sendPoll"; see %s`, greenAPIDocsURL)
	}
	httpMethod, _ = obj["httpMethod"].(string)
	if httpMethod == "" {
		return "", "", nil, fmt.Errorf(`request.httpMethod is required (GET, POST or DELETE, as documented for %q); see %s`, method, greenAPIDocsURL)
	}
	body = obj["body"]
	return method, httpMethod, body, nil
}

func rawResult(result json.RawMessage, err error) (*mcp.CallToolResult, error) {
	return rawResultOr(`{}`, result, err)
}

func rawResultOr(emptyResult string, result json.RawMessage, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if len(result) == 0 {
		return mcp.NewToolResultText(emptyResult), nil
	}
	return mcp.NewToolResultText(string(result)), nil
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

func requireStrings(req mcp.CallToolRequest, keys ...string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		v, err := req.RequireString(key)
		if err != nil {
			return nil, err
		}
		values[key] = v
	}
	return values, nil
}

func decodeArg(raw any, target any) error {
	data, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func jsonStringArg(args map[string]any, key string) (string, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return "", nil
	}
	if str, ok := raw.(string); ok {
		return str, nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return "", fmt.Errorf("%s: %w", key, err)
	}
	return string(data), nil
}

func templateMediaMessage(mediaType, link, filename string) *domain.TemplateMessage {
	media := &domain.TemplateMedia{Link: link, Filename: filename}
	msg := &domain.TemplateMessage{Type: mediaType}
	switch mediaType {
	case domain.TemplateMediaImage:
		msg.Image = media
	case domain.TemplateMediaVideo:
		msg.Video = media
	case domain.TemplateMediaDocument:
		msg.Document = media
	}
	return msg
}

func yesNo(v bool) string {
	if v {
		return domain.SettingEnabled
	}
	return domain.SettingDisabled
}

func clampInt(v, fallback, upper int) int {
	if v < 1 {
		return fallback
	}
	if v > upper {
		return upper
	}
	return v
}

func resolveInstanceID(ctx context.Context, req mcp.CallToolRequest) (uint64, error) {
	if _, ok := req.GetArguments()["instance_id"]; ok {
		f, err := req.RequireFloat("instance_id")
		if err != nil {
			return 0, err
		}
		return uint64(f), nil
	}
	if id, ok := infrastructure.GetInstanceIDFromContext(ctx); ok && id > 0 {
		return id, nil
	}
	return 0, fmt.Errorf("instance_id is required")
}
