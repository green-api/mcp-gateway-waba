package domain

import (
	"encoding/json"
	"testing"
)

func marshalToMap(t *testing.T, v any) map[string]any {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m
}

func TestSendMessageRequest_LinkPreviewOmittedWhenUnset(t *testing.T) {
	m := marshalToMap(t, SendMessageRequest{ChatID: "11001234567@c.us", Message: "Hi"})
	if _, ok := m["linkPreview"]; ok {
		t.Error("linkPreview should be omitted when nil")
	}
	if m["chatId"] != "11001234567@c.us" || m["message"] != "Hi" {
		t.Errorf("unexpected body: %v", m)
	}
}

func TestSendMessageRequest_LinkPreviewFalseIsSent(t *testing.T) {
	off := false
	m := marshalToMap(t, SendMessageRequest{ChatID: "11001234567@c.us", Message: "Hi", LinkPreview: &off})
	if v, ok := m["linkPreview"]; !ok || v != false {
		t.Errorf("linkPreview: got %v, want false", v)
	}
}

func TestSendFileByURLRequest_FileNameAlwaysPresent(t *testing.T) {
	m := marshalToMap(t, SendFileByURLRequest{ChatID: "11001234567@c.us", URL: "https://x/y.png", FileName: "y.png"})
	if m["urlFile"] != "https://x/y.png" {
		t.Errorf("urlFile: got %v", m["urlFile"])
	}
	if m["fileName"] != "y.png" {
		t.Errorf("fileName: got %v", m["fileName"])
	}
	if _, ok := m["caption"]; ok {
		t.Error("caption should be omitted when empty")
	}
}

func TestSendTemplateRequest_JSON(t *testing.T) {
	req := SendTemplateRequest{
		ChatID:     "11001234567@c.us",
		TemplateID: "tpl-1",
		Params:     []string{"John", "15"},
		Message: &TemplateMessage{
			Type:  TemplateMediaImage,
			Image: &TemplateMedia{Link: "https://x/img.jpg"},
		},
		PostbackTexts: []PostbackText{{Index: 0, Text: "yes"}},
	}
	m := marshalToMap(t, req)
	if m["templateId"] != "tpl-1" {
		t.Errorf("templateId: got %v", m["templateId"])
	}
	params, ok := m["params"].([]any)
	if !ok || len(params) != 2 || params[0] != "John" {
		t.Errorf("params: got %v", m["params"])
	}
	msg, ok := m["message"].(map[string]any)
	if !ok || msg["type"] != "image" {
		t.Fatalf("message: got %v", m["message"])
	}
	img, ok := msg["image"].(map[string]any)
	if !ok || img["link"] != "https://x/img.jpg" {
		t.Errorf("message.image: got %v", msg["image"])
	}
	if _, ok := img["filename"]; ok {
		t.Error("image.filename should be omitted when empty")
	}
	if _, ok := msg["video"]; ok {
		t.Error("video should be omitted")
	}
	pb, ok := m["postbackTexts"].([]any)
	if !ok || len(pb) != 1 {
		t.Errorf("postbackTexts: got %v", m["postbackTexts"])
	}
}

func TestSendTemplateRequest_MinimalOmitsOptional(t *testing.T) {
	m := marshalToMap(t, SendTemplateRequest{ChatID: "11001234567@c.us", TemplateID: "tpl-1"})
	for _, key := range []string{"params", "message", "postbackTexts"} {
		if _, ok := m[key]; ok {
			t.Errorf("%s should be omitted when empty", key)
		}
	}
}

func TestCreateTemplateRequest_JSON(t *testing.T) {
	req := CreateTemplateRequest{
		ElementName:  "order_update",
		LanguageCode: "en_US",
		Category:     TemplateCategoryUtility,
		TemplateType: TemplateTypeText,
		Vertical:     "Order updates",
		Content:      "Your order {{1}} is ready",
		Example:      "Your order 42 is ready",
		Buttons:      `[{"type":"QUICK_REPLY","text":"Thanks"}]`,
		EnableSample: true,
	}
	m := marshalToMap(t, req)
	if m["elementName"] != "order_update" || m["category"] != "UTILITY" || m["templateType"] != "TEXT" {
		t.Errorf("unexpected body: %v", m)
	}
	if v, ok := m["buttons"].(string); !ok || v == "" {
		t.Errorf("buttons must be serialized as a JSON string, got %T %v", m["buttons"], m["buttons"])
	}
	if m["enableSample"] != true {
		t.Errorf("enableSample: got %v", m["enableSample"])
	}
	for _, key := range []string{"header", "footer", "cards", "mediaUrl", "codeExpirationMinutes", "allowTemplateCategoryChange"} {
		if _, ok := m[key]; ok {
			t.Errorf("%s should be omitted when empty", key)
		}
	}
}

func TestCreateTemplateRequest_EnableSampleFalseIsSent(t *testing.T) {
	m := marshalToMap(t, CreateTemplateRequest{ElementName: "x", EnableSample: false})
	if v, ok := m["enableSample"]; !ok || v != false {
		t.Errorf("enableSample: got %v", v)
	}
}

func TestEditTemplateRequest_JSON(t *testing.T) {
	on := true
	m := marshalToMap(t, EditTemplateRequest{
		TemplateID:     "tpl-1",
		TemplateParams: EditTemplateParams{Content: "new body", EnableSample: &on},
	})
	if m["templateId"] != "tpl-1" {
		t.Errorf("templateId: got %v", m["templateId"])
	}
	params, ok := m["templateParams"].(map[string]any)
	if !ok {
		t.Fatalf("templateParams: got %v", m["templateParams"])
	}
	if params["content"] != "new body" || params["enableSample"] != true {
		t.Errorf("templateParams: got %v", params)
	}
	if _, ok := params["header"]; ok {
		t.Error("header should be omitted when empty")
	}
}

func TestDeleteTemplateRequests_JSON(t *testing.T) {
	m := marshalToMap(t, DeleteTemplateRequest{ElementName: "order_update"})
	if m["elementName"] != "order_update" {
		t.Errorf("elementName: got %v", m["elementName"])
	}
	m = marshalToMap(t, DeleteTemplateByIDRequest{ElementName: "order_update", TemplateID: "tpl-1"})
	if m["elementName"] != "order_update" || m["templateId"] != "tpl-1" {
		t.Errorf("unexpected body: %v", m)
	}
	m = marshalToMap(t, GetTemplateByIDRequest{TemplateID: "tpl-1"})
	if m["templateId"] != "tpl-1" {
		t.Errorf("templateId: got %v", m["templateId"])
	}
}

func TestGetTemplatesResponse_Unmarshal(t *testing.T) {
	raw := `{"templates":[{"templateId":"tpl-1","elementName":"order_update","status":"APPROVED","category":"UTILITY","templateType":"TEXT","languageCode":"en_US","data":"Your order {{1}} is ready","createdOn":1741089781669,"modifiedOn":1741089798570}]}`
	var resp GetTemplatesResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Templates) != 1 {
		t.Fatalf("templates: got %d", len(resp.Templates))
	}
	tpl := resp.Templates[0]
	if tpl.TemplateID != "tpl-1" || tpl.Status != TemplateStatusApproved || tpl.ModifiedOn != 1741089798570 {
		t.Errorf("unexpected template: %+v", tpl)
	}
}

func TestTemplateResponse_Unmarshal(t *testing.T) {
	raw := `{"template":{"templateId":"tpl-1","elementName":"x","status":"PENDING","category":"MARKETING","templateType":"IMAGE","languageCode":"ru"}}`
	var resp TemplateResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Template.Status != TemplateStatusPending || resp.Template.TemplateType != TemplateTypeImage {
		t.Errorf("unexpected template: %+v", resp.Template)
	}
}

func TestTemplateStatusResponse_Unmarshal(t *testing.T) {
	var resp TemplateStatusResponse
	if err := json.Unmarshal([]byte(`{"status":"failed","reason":"Only Rejected, Approved and Paused templates can be edited"}`), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Status != "failed" || resp.Reason == "" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestSetSettingsRequest_EmptyWebhookURLIsSent(t *testing.T) {
	empty := ""
	m := marshalToMap(t, SetSettingsRequest{WebhookURL: &empty})
	if v, ok := m["webhookUrl"]; !ok || v != "" {
		t.Errorf("webhookUrl: got %v, want empty string", v)
	}
	if _, ok := m["webhookUrlToken"]; ok {
		t.Error("webhookUrlToken should be omitted when nil")
	}
}

func TestSetSettingsRequest_WebhookFlags(t *testing.T) {
	m := marshalToMap(t, SetSettingsRequest{OutgoingWebhook: SettingEnabled, IncomingWebhook: SettingDisabled})
	if m["outgoingWebhook"] != "yes" || m["incomingWebhook"] != "no" {
		t.Errorf("unexpected body: %v", m)
	}
	if _, ok := m["outgoingAPIMessageWebhook"]; ok {
		t.Error("outgoingAPIMessageWebhook should be omitted when empty")
	}
}

func TestSettingsResponse_Unmarshal(t *testing.T) {
	raw := `{"wid":"11001234567@c.us","webhookUrl":"https://mysite.com/webhook/","webhookUrlToken":"","outgoingWebhook":"yes","outgoingAPIMessageWebhook":"yes","incomingWebhook":"yes","deviceWebhook":"no"}`
	var resp SettingsResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Wid != "11001234567@c.us" || resp.WebhookURL != "https://mysite.com/webhook/" || resp.IncomingWebhook != "yes" {
		t.Errorf("unexpected settings: %+v", resp)
	}
}

func TestSetSettingsResponse_Unmarshal(t *testing.T) {
	var resp SetSettingsResponse
	if err := json.Unmarshal([]byte(`{"saveSettings":true}`), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.SaveSettings {
		t.Error("SaveSettings: expected true")
	}
}

func TestWaSettingsResponse_Unmarshal(t *testing.T) {
	var resp WaSettingsResponse
	if err := json.Unmarshal([]byte(`{"avatar":"","phone":"79876543210","stateInstance":"authorized","deviceId":""}`), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Phone != "79876543210" || resp.StateInstance != "authorized" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestSendFileByUploadResponse_Unmarshal(t *testing.T) {
	var resp SendFileByUploadResponse
	if err := json.Unmarshal([]byte(`{"idMessage":"3EB0","urlFile":"https://storage/x.jpg"}`), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.IDMessage != "3EB0" || resp.URLFile != "https://storage/x.jpg" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestNotificationBody_KeepsRawBody(t *testing.T) {
	raw := `{"receiptId":42,"body":{"typeWebhook":"incomingMessageReceived","messageData":{"typeMessage":"textMessage"}}}`
	var n NotificationBody
	if err := json.Unmarshal([]byte(raw), &n); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if n.ReceiptID != 42 {
		t.Errorf("ReceiptID: got %d", n.ReceiptID)
	}
	var body map[string]any
	if err := json.Unmarshal(n.Body, &body); err != nil {
		t.Fatalf("body: %v", err)
	}
	if body["typeWebhook"] != "incomingMessageReceived" {
		t.Errorf("typeWebhook: got %v", body["typeWebhook"])
	}
}

func TestJournalRequests_JSON(t *testing.T) {
	m := marshalToMap(t, GetChatHistoryRequest{ChatID: "11001234567@c.us", Count: 20})
	if m["chatId"] != "11001234567@c.us" || m["count"] != float64(20) {
		t.Errorf("unexpected body: %v", m)
	}
	m = marshalToMap(t, GetChatHistoryRequest{ChatID: "11001234567@c.us"})
	if _, ok := m["count"]; ok {
		t.Error("count should be omitted when zero")
	}
	m = marshalToMap(t, GetMessageRequest{ChatID: "11001234567@c.us", IDMessage: "abc"})
	if m["idMessage"] != "abc" {
		t.Errorf("idMessage: got %v", m["idMessage"])
	}
}
