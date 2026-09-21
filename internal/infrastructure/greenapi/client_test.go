package greenapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/green-api/green-api-mcp-gateway-waba/internal/domain"
	"github.com/green-api/green-api-mcp-gateway-waba/internal/infrastructure"
	"github.com/green-api/green-api-mcp-gateway-waba/internal/infrastructure/greenapi"
	"github.com/green-api/green-api-mcp-gateway-waba/internal/infrastructure/monitoring"
)

const (
	testInstanceID uint64 = 1234567890
	testToken             = "test-token"
)

func newClientWithHandler(t *testing.T, handler http.HandlerFunc) (*greenapi.Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)

	mgr := infrastructure.NewCredentialManager()
	mgr.AddInstance(&domain.InstanceCredentials{
		InstanceID: testInstanceID,
		APIToken:   testToken,
		APIURL:     srv.URL,
	})

	client := greenapi.NewClient(mgr, monitoring.NoopMetricsProvider{}, nil)
	return client, srv.Close
}

func newClientWithServer(t *testing.T, statusCode int, body string) (*greenapi.Client, func()) {
	t.Helper()
	return newClientWithHandler(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	})
}

type recorded struct {
	method string
	path   string
	query  string
	body   map[string]any
}

func newRecordingClient(t *testing.T, responseBody string, rec *recorded) (*greenapi.Client, func()) {
	t.Helper()
	return newClientWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		rec.method = r.Method
		rec.path = r.URL.Path
		rec.query = r.URL.RawQuery
		rec.body = nil
		data, _ := io.ReadAll(r.Body)
		if len(data) > 0 {
			_ = json.Unmarshal(data, &rec.body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseBody))
	})
}

func methodPath(method string) string {
	return fmt.Sprintf("/waInstance%d/%s/%s", testInstanceID, method, testToken)
}

func TestClient_UnknownInstance_ReturnsError(t *testing.T) {
	mgr := infrastructure.NewCredentialManager()
	client := greenapi.NewClient(mgr, monitoring.NoopMetricsProvider{}, nil)

	if _, err := client.GetStateInstance(context.Background(), 99999); err == nil {
		t.Fatal("expected error for unknown instance, got nil")
	}
}

func TestClient_CallMethod_Success(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"idMessage":"msg-001"}`, rec)
	defer cleanup()

	raw, err := client.CallMethod(context.Background(), testInstanceID, domain.MethodSendMessage, map[string]any{
		"chatId":  "11001234567@c.us",
		"message": "hello",
	})
	if err != nil {
		t.Fatalf("CallMethod: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["idMessage"] != "msg-001" {
		t.Errorf("idMessage: got %v", got["idMessage"])
	}
	if rec.method != http.MethodPost || rec.path != methodPath(domain.MethodSendMessage) {
		t.Errorf("request: %s %s", rec.method, rec.path)
	}
}

func TestClient_CallMethod_ServerError(t *testing.T) {
	client, cleanup := newClientWithServer(t, http.StatusInternalServerError, "server error")
	defer cleanup()

	if _, err := client.CallMethod(context.Background(), testInstanceID, "someMethod", nil); err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestClient_VerbMapping(t *testing.T) {
	tests := []struct {
		method string
		verb   string
	}{
		{domain.MethodGetTemplates, http.MethodGet},
		{domain.MethodGetStateInstance, http.MethodGet},
		{domain.MethodGetWaSettings, http.MethodGet},
		{domain.MethodReboot, http.MethodGet},
		{domain.MethodShowMessagesQueue, http.MethodGet},
		{domain.MethodClearMessagesQueue, http.MethodGet},
		{domain.MethodReceiveNotification, http.MethodGet},
		{domain.MethodDeleteNotification, http.MethodDelete},
		{domain.MethodGetTemplateByID, http.MethodPost},
		{domain.MethodCreateTemplate, http.MethodPost},
		{domain.MethodSendTemplate, http.MethodPost},
		{domain.MethodSetSettings, http.MethodPost},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			rec := &recorded{}
			client, cleanup := newRecordingClient(t, `{}`, rec)
			defer cleanup()

			if _, err := client.CallMethod(context.Background(), testInstanceID, tt.method, nil); err != nil {
				t.Fatalf("CallMethod: %v", err)
			}
			if rec.method != tt.verb {
				t.Errorf("%s: verb %s, want %s", tt.method, rec.method, tt.verb)
			}
		})
	}
}

func TestClient_GetMethod_Success(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"stateInstance":"authorized"}`, rec)
	defer cleanup()

	raw, err := client.GetMethod(context.Background(), testInstanceID, domain.MethodGetStateInstance)
	if err != nil {
		t.Fatalf("GetMethod: %v", err)
	}
	if rec.method != http.MethodGet {
		t.Errorf("verb: %s", rec.method)
	}
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	if got["stateInstance"] != "authorized" {
		t.Errorf("stateInstance: got %v", got["stateInstance"])
	}
}

func TestClient_SendMessage_RequestBody(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"idMessage":"abc-123"}`, rec)
	defer cleanup()

	resp, err := client.SendMessage(context.Background(), testInstanceID, domain.SendMessageRequest{
		ChatID:  "11001234567@c.us",
		Message: "Hello world",
	})
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if resp.IDMessage != "abc-123" {
		t.Errorf("IDMessage: got %q", resp.IDMessage)
	}
	if rec.body["chatId"] != "11001234567@c.us" || rec.body["message"] != "Hello world" {
		t.Errorf("body: %v", rec.body)
	}
	if _, ok := rec.body["linkPreview"]; ok {
		t.Error("linkPreview must be omitted when unset")
	}
}

func TestClient_SendMessage_Error(t *testing.T) {
	client, cleanup := newClientWithServer(t, http.StatusBadRequest, `{"error":"bad request"}`)
	defer cleanup()

	if _, err := client.SendMessage(context.Background(), testInstanceID, domain.SendMessageRequest{ChatID: "x@c.us", Message: "t"}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_SendTemplate_RequestBody(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"idMessage":"tpl-msg"}`, rec)
	defer cleanup()

	resp, err := client.SendTemplate(context.Background(), testInstanceID, domain.SendTemplateRequest{
		ChatID:     "11001234567@c.us",
		TemplateID: "2522g44c",
		Params:     []string{"John", "15"},
		Message:    &domain.TemplateMessage{Type: domain.TemplateMediaImage, Image: &domain.TemplateMedia{Link: "https://x/img.jpg"}},
	})
	if err != nil {
		t.Fatalf("SendTemplate: %v", err)
	}
	if resp.IDMessage != "tpl-msg" {
		t.Errorf("IDMessage: got %q", resp.IDMessage)
	}
	if rec.path != methodPath(domain.MethodSendTemplate) || rec.method != http.MethodPost {
		t.Errorf("request: %s %s", rec.method, rec.path)
	}
	if rec.body["templateId"] != "2522g44c" {
		t.Errorf("templateId: got %v", rec.body["templateId"])
	}
	params, _ := rec.body["params"].([]any)
	if len(params) != 2 {
		t.Errorf("params: got %v", rec.body["params"])
	}
	msg, _ := rec.body["message"].(map[string]any)
	if msg["type"] != "image" {
		t.Errorf("message: got %v", rec.body["message"])
	}
}

func TestClient_SendFileByURL_Success(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"idMessage":"file-456"}`, rec)
	defer cleanup()

	resp, err := client.SendFileByURL(context.Background(), testInstanceID, domain.SendFileByURLRequest{
		ChatID:   "11001234567@c.us",
		URL:      "https://example.com/file.pdf",
		FileName: "file.pdf",
	})
	if err != nil {
		t.Fatalf("SendFileByURL: %v", err)
	}
	if resp.IDMessage != "file-456" {
		t.Errorf("IDMessage: got %q", resp.IDMessage)
	}
	if rec.body["urlFile"] != "https://example.com/file.pdf" || rec.body["fileName"] != "file.pdf" {
		t.Errorf("body: %v", rec.body)
	}
}

func TestClient_SendFileByUpload_Multipart(t *testing.T) {
	var gotChatID, gotFileName, gotCaption, gotFile string
	client, cleanup := newClientWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("multipart: %v", err)
		}
		gotChatID = r.FormValue("chatId")
		gotFileName = r.FormValue("fileName")
		gotCaption = r.FormValue("caption")
		f, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("file: %v", err)
		} else {
			data, _ := io.ReadAll(f)
			gotFile = string(data)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"idMessage":"up-1","urlFile":"https://storage/x.txt"}`))
	})
	defer cleanup()

	resp, err := client.SendFileByUpload(context.Background(), testInstanceID, "11001234567@c.us", "hi", "x.txt", []byte("payload"))
	if err != nil {
		t.Fatalf("SendFileByUpload: %v", err)
	}
	if resp.IDMessage != "up-1" || resp.URLFile != "https://storage/x.txt" {
		t.Errorf("response: %+v", resp)
	}
	if gotChatID != "11001234567@c.us" || gotFileName != "x.txt" || gotCaption != "hi" || gotFile != "payload" {
		t.Errorf("form: chatId=%q fileName=%q caption=%q file=%q", gotChatID, gotFileName, gotCaption, gotFile)
	}
}

func TestClient_GetStateInstance(t *testing.T) {
	client, cleanup := newClientWithServer(t, http.StatusOK, `{"stateInstance":"notAuthorized"}`)
	defer cleanup()

	resp, err := client.GetStateInstance(context.Background(), testInstanceID)
	if err != nil {
		t.Fatalf("GetStateInstance: %v", err)
	}
	if resp.StateInstance != "notAuthorized" {
		t.Errorf("StateInstance: got %q", resp.StateInstance)
	}
}

func TestClient_GetSettings_Success(t *testing.T) {
	client, cleanup := newClientWithServer(t, http.StatusOK, `{"wid":"11001234567@c.us","webhookUrl":"https://example.com/hook","incomingWebhook":"yes"}`)
	defer cleanup()

	resp, err := client.GetSettings(context.Background(), testInstanceID)
	if err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	if resp.Wid != "11001234567@c.us" || resp.WebhookURL != "https://example.com/hook" || resp.IncomingWebhook != "yes" {
		t.Errorf("settings: %+v", resp)
	}
}

func TestClient_SetSettings_Success(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"saveSettings":true}`, rec)
	defer cleanup()

	empty := ""
	resp, err := client.SetSettings(context.Background(), testInstanceID, domain.SetSettingsRequest{
		WebhookURL:      &empty,
		IncomingWebhook: domain.SettingEnabled,
	})
	if err != nil {
		t.Fatalf("SetSettings: %v", err)
	}
	if !resp.SaveSettings {
		t.Error("SaveSettings: expected true")
	}
	if v, ok := rec.body["webhookUrl"]; !ok || v != "" {
		t.Errorf("webhookUrl: got %v, want empty string", v)
	}
	if rec.body["incomingWebhook"] != "yes" {
		t.Errorf("incomingWebhook: got %v", rec.body["incomingWebhook"])
	}
}

func TestClient_GetWaSettings_Success(t *testing.T) {
	client, cleanup := newClientWithServer(t, http.StatusOK, `{"avatar":"","phone":"79876543210","stateInstance":"authorized","deviceId":""}`)
	defer cleanup()

	resp, err := client.GetWaSettings(context.Background(), testInstanceID)
	if err != nil {
		t.Fatalf("GetWaSettings: %v", err)
	}
	if resp.Phone != "79876543210" {
		t.Errorf("Phone: got %q", resp.Phone)
	}
}

func TestClient_Reboot_Success(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"isReboot":true}`, rec)
	defer cleanup()

	raw, err := client.Reboot(context.Background(), testInstanceID)
	if err != nil {
		t.Fatalf("Reboot: %v", err)
	}
	if rec.method != http.MethodGet || string(raw) != `{"isReboot":true}` {
		t.Errorf("request: %s, body %s", rec.method, raw)
	}
}

func TestClient_CreateTemplate_Success(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"template":{"templateId":"tpl-1","elementName":"order_update","status":"PENDING","category":"UTILITY","templateType":"TEXT","languageCode":"en_US"}}`, rec)
	defer cleanup()

	resp, err := client.CreateTemplate(context.Background(), testInstanceID, domain.CreateTemplateRequest{
		ElementName:  "order_update",
		LanguageCode: "en_US",
		Category:     domain.TemplateCategoryUtility,
		TemplateType: domain.TemplateTypeText,
		Vertical:     "Order updates",
		Content:      "Your order {{1}} is ready",
		Example:      "Your order 42 is ready",
		EnableSample: true,
	})
	if err != nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	if resp.Template.TemplateID != "tpl-1" || resp.Template.Status != domain.TemplateStatusPending {
		t.Errorf("template: %+v", resp.Template)
	}
	if rec.method != http.MethodPost || rec.path != methodPath(domain.MethodCreateTemplate) {
		t.Errorf("request: %s %s", rec.method, rec.path)
	}
	if rec.body["elementName"] != "order_update" || rec.body["enableSample"] != true {
		t.Errorf("body: %v", rec.body)
	}
}

func TestClient_EditTemplate_Failed(t *testing.T) {
	client, cleanup := newClientWithServer(t, http.StatusOK, `{"status":"failed","reason":"Only Rejected, Approved and Paused templates can be edited"}`)
	defer cleanup()

	resp, err := client.EditTemplate(context.Background(), testInstanceID, domain.EditTemplateRequest{
		TemplateID:     "tpl-1",
		TemplateParams: domain.EditTemplateParams{Content: "new"},
	})
	if err != nil {
		t.Fatalf("EditTemplate: %v", err)
	}
	if resp.Status != "failed" || resp.Reason == "" {
		t.Errorf("response: %+v", resp)
	}
}

func TestClient_GetTemplates_Success(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"templates":[{"templateId":"a","elementName":"one","status":"APPROVED","category":"MARKETING","templateType":"TEXT","languageCode":"en"},{"templateId":"b","elementName":"two","status":"PENDING","category":"UTILITY","templateType":"IMAGE","languageCode":"ru"}]}`, rec)
	defer cleanup()

	resp, err := client.GetTemplates(context.Background(), testInstanceID)
	if err != nil {
		t.Fatalf("GetTemplates: %v", err)
	}
	if rec.method != http.MethodGet || rec.path != methodPath(domain.MethodGetTemplates) {
		t.Errorf("request: %s %s", rec.method, rec.path)
	}
	if len(resp.Templates) != 2 || resp.Templates[1].Status != domain.TemplateStatusPending {
		t.Errorf("templates: %+v", resp.Templates)
	}
}

func TestClient_GetTemplateByID_PostsBody(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"template":{"templateId":"tpl-1","elementName":"one","status":"APPROVED","category":"UTILITY","templateType":"TEXT","languageCode":"en"}}`, rec)
	defer cleanup()

	resp, err := client.GetTemplateByID(context.Background(), testInstanceID, domain.GetTemplateByIDRequest{TemplateID: "tpl-1"})
	if err != nil {
		t.Fatalf("GetTemplateByID: %v", err)
	}
	if rec.method != http.MethodPost || rec.body["templateId"] != "tpl-1" {
		t.Errorf("request: %s body %v", rec.method, rec.body)
	}
	if resp.Template.ElementName != "one" {
		t.Errorf("template: %+v", resp.Template)
	}
}

func TestClient_DeleteTemplate_Success(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"status":"success"}`, rec)
	defer cleanup()

	resp, err := client.DeleteTemplate(context.Background(), testInstanceID, domain.DeleteTemplateRequest{ElementName: "one"})
	if err != nil {
		t.Fatalf("DeleteTemplate: %v", err)
	}
	if resp.Status != "success" || rec.body["elementName"] != "one" || rec.path != methodPath(domain.MethodDeleteTemplate) {
		t.Errorf("response %+v, body %v, path %s", resp, rec.body, rec.path)
	}
}

func TestClient_DeleteTemplateByID_Success(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"status":"success"}`, rec)
	defer cleanup()

	_, err := client.DeleteTemplateByID(context.Background(), testInstanceID, domain.DeleteTemplateByIDRequest{ElementName: "one", TemplateID: "tpl-1"})
	if err != nil {
		t.Fatalf("DeleteTemplateByID: %v", err)
	}
	if rec.body["templateId"] != "tpl-1" || rec.body["elementName"] != "one" || rec.path != methodPath(domain.MethodDeleteTemplateByID) {
		t.Errorf("body %v, path %s", rec.body, rec.path)
	}
}

func TestClient_ReceiveNotification_HasNotification(t *testing.T) {
	client, cleanup := newClientWithServer(t, http.StatusOK, `{"receiptId":42,"body":{"typeWebhook":"incomingMessageReceived"}}`)
	defer cleanup()

	notif, err := client.ReceiveNotification(context.Background(), testInstanceID)
	if err != nil {
		t.Fatalf("ReceiveNotification: %v", err)
	}
	if notif == nil || notif.ReceiptID != 42 {
		t.Fatalf("notification: %+v", notif)
	}
}

func TestClient_ReceiveNotification_Empty(t *testing.T) {
	client, cleanup := newClientWithServer(t, http.StatusOK, ``)
	defer cleanup()

	notif, err := client.ReceiveNotification(context.Background(), testInstanceID)
	if err != nil {
		t.Fatalf("ReceiveNotification: %v", err)
	}
	if notif != nil {
		t.Errorf("expected nil notification, got %+v", notif)
	}
}

func TestClient_DeleteNotification_PathAndVerb(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"result":true}`, rec)
	defer cleanup()

	if err := client.DeleteNotification(context.Background(), testInstanceID, 42); err != nil {
		t.Fatalf("DeleteNotification: %v", err)
	}
	want := fmt.Sprintf("/waInstance%d/deleteNotification/42/%s", testInstanceID, testToken)
	if rec.method != http.MethodDelete || rec.path != want {
		t.Errorf("request: %s %s, want DELETE %s", rec.method, rec.path, want)
	}
}

func TestClient_DeleteNotification_Error(t *testing.T) {
	client, cleanup := newClientWithServer(t, http.StatusNotFound, `{"error":"not found"}`)
	defer cleanup()

	if err := client.DeleteNotification(context.Background(), testInstanceID, 99); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_GetChatHistory_Success(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `[{"idMessage":"m1"}]`, rec)
	defer cleanup()

	raw, err := client.GetChatHistory(context.Background(), testInstanceID, domain.GetChatHistoryRequest{ChatID: "11001234567@c.us", Count: 10})
	if err != nil {
		t.Fatalf("GetChatHistory: %v", err)
	}
	if rec.method != http.MethodPost || rec.body["count"] != float64(10) {
		t.Errorf("request: %s body %v", rec.method, rec.body)
	}
	var msgs []map[string]any
	if err := json.Unmarshal(raw, &msgs); err != nil || len(msgs) != 1 {
		t.Errorf("history: %s (%v)", raw, err)
	}
}

func TestClient_GetMessage_Success(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"idMessage":"m1","textMessage":"hi"}`, rec)
	defer cleanup()

	if _, err := client.GetMessage(context.Background(), testInstanceID, domain.GetMessageRequest{ChatID: "11001234567@c.us", IDMessage: "m1"}); err != nil {
		t.Fatalf("GetMessage: %v", err)
	}
	if rec.body["idMessage"] != "m1" {
		t.Errorf("body: %v", rec.body)
	}
}

func TestClient_LastIncomingMessages_QueryAndVerb(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `[]`, rec)
	defer cleanup()

	if _, err := client.LastIncomingMessages(context.Background(), testInstanceID, 30); err != nil {
		t.Fatalf("LastIncomingMessages: %v", err)
	}
	if rec.method != http.MethodGet || rec.path != methodPath(domain.MethodLastIncomingMessages) || rec.query != "minutes=30" {
		t.Errorf("request: %s %s?%s", rec.method, rec.path, rec.query)
	}
}

func TestClient_LastOutgoingMessages_QueryAndVerb(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `[]`, rec)
	defer cleanup()

	if _, err := client.LastOutgoingMessages(context.Background(), testInstanceID, 1440); err != nil {
		t.Fatalf("LastOutgoingMessages: %v", err)
	}
	if rec.method != http.MethodGet || rec.query != "minutes=1440" {
		t.Errorf("request: %s ?%s", rec.method, rec.query)
	}
}

func TestClient_RawMethod_GET(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `[{"id":"1@c.us"}]`, rec)
	defer cleanup()

	raw, err := client.RawMethod(context.Background(), testInstanceID, "GET", "getChats", nil)
	if err != nil {
		t.Fatalf("RawMethod: %v", err)
	}
	if rec.method != http.MethodGet || rec.path != methodPath("getChats") {
		t.Errorf("request: %s %s, want GET %s", rec.method, rec.path, methodPath("getChats"))
	}
	var chats []map[string]any
	if err := json.Unmarshal(raw, &chats); err != nil || len(chats) != 1 {
		t.Errorf("response: %s (%v)", raw, err)
	}
}

func TestClient_RawMethod_DELETE(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{}`, rec)
	defer cleanup()

	if _, err := client.RawMethod(context.Background(), testInstanceID, "delete", "clearWebhooksQueue", nil); err != nil {
		t.Fatalf("RawMethod: %v", err)
	}
	if rec.method != http.MethodDelete || rec.path != methodPath("clearWebhooksQueue") {
		t.Errorf("request: %s %s", rec.method, rec.path)
	}
}

func TestClient_RawMethod_POST_ForwardsBody(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `{"created":true,"chatId":"123-456@g.us"}`, rec)
	defer cleanup()

	raw, err := client.RawMethod(context.Background(), testInstanceID, "post", "createGroup", map[string]any{
		"groupName": "Test",
		"chatIds":   []string{"11001234567@c.us"},
	})
	if err != nil {
		t.Fatalf("RawMethod: %v", err)
	}
	if rec.method != http.MethodPost || rec.path != methodPath("createGroup") {
		t.Errorf("request: %s %s", rec.method, rec.path)
	}
	if rec.body["groupName"] != "Test" {
		t.Errorf("body: %v", rec.body)
	}
	var resp map[string]any
	if err := json.Unmarshal(raw, &resp); err != nil || resp["chatId"] != "123-456@g.us" {
		t.Errorf("response: %s (%v)", raw, err)
	}
}

func TestClient_RawMethod_RejectsMissingVerb(t *testing.T) {
	mgr := infrastructure.NewCredentialManager()
	mgr.AddInstance(&domain.InstanceCredentials{InstanceID: testInstanceID, APIToken: testToken, APIURL: "http://example.invalid"})
	client := greenapi.NewClient(mgr, monitoring.NoopMetricsProvider{}, nil)

	// No implicit default: an empty httpMethod is rejected, same as any other unsupported verb.
	if _, err := client.RawMethod(context.Background(), testInstanceID, "", "getChats", nil); err == nil {
		t.Fatal("expected error for empty http method")
	}
}

func TestClient_RawMethod_RejectsUnsupportedVerb(t *testing.T) {
	mgr := infrastructure.NewCredentialManager()
	mgr.AddInstance(&domain.InstanceCredentials{InstanceID: testInstanceID, APIToken: testToken, APIURL: "http://example.invalid"})
	client := greenapi.NewClient(mgr, monitoring.NoopMetricsProvider{}, nil)

	if _, err := client.RawMethod(context.Background(), testInstanceID, "PATCH", "getChats", nil); err == nil {
		t.Fatal("expected error for unsupported http method")
	}
}

func TestClient_RawMethod_RejectsBodyWithNonPOSTVerb(t *testing.T) {
	mgr := infrastructure.NewCredentialManager()
	mgr.AddInstance(&domain.InstanceCredentials{InstanceID: testInstanceID, APIToken: testToken, APIURL: "http://example.invalid"})
	client := greenapi.NewClient(mgr, monitoring.NoopMetricsProvider{}, nil)

	if _, err := client.RawMethod(context.Background(), testInstanceID, "GET", "getChats", map[string]any{"x": 1}); err == nil {
		t.Fatal("expected error for body with a non-POST verb")
	}
}

func TestClient_RawMethod_ServerError(t *testing.T) {
	client, cleanup := newClientWithServer(t, http.StatusForbidden, `{"error":"forbidden"}`)
	defer cleanup()

	if _, err := client.RawMethod(context.Background(), testInstanceID, "GET", "getChats", nil); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_MessagesQueue(t *testing.T) {
	rec := &recorded{}
	client, cleanup := newRecordingClient(t, `[]`, rec)
	defer cleanup()

	if _, err := client.ShowMessagesQueue(context.Background(), testInstanceID); err != nil {
		t.Fatalf("ShowMessagesQueue: %v", err)
	}
	if rec.method != http.MethodGet || rec.path != methodPath(domain.MethodShowMessagesQueue) {
		t.Errorf("request: %s %s", rec.method, rec.path)
	}

	if _, err := client.ClearMessagesQueue(context.Background(), testInstanceID); err != nil {
		t.Fatalf("ClearMessagesQueue: %v", err)
	}
	if rec.method != http.MethodGet || rec.path != methodPath(domain.MethodClearMessagesQueue) {
		t.Errorf("request: %s %s", rec.method, rec.path)
	}
}
