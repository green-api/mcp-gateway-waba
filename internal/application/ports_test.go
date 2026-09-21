package application_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/green-api/green-api-mcp-gateway-waba/internal/application"
	"github.com/green-api/green-api-mcp-gateway-waba/internal/domain"
)

type mockCredentialStore struct {
	instances map[uint64]*domain.InstanceCredentials
}

func newMockCredentialStore() *mockCredentialStore {
	return &mockCredentialStore{instances: make(map[uint64]*domain.InstanceCredentials)}
}

func (m *mockCredentialStore) AddInstance(creds *domain.InstanceCredentials) {
	m.instances[creds.InstanceID] = creds
}

func (m *mockCredentialStore) GetCredentials(id uint64) (*domain.InstanceCredentials, error) {
	c, ok := m.instances[id]
	if !ok {
		return nil, domain.ErrInstanceNotFound
	}
	return c, nil
}

func (m *mockCredentialStore) ListInstances() []uint64 {
	ids := make([]uint64, 0, len(m.instances))
	for id := range m.instances {
		ids = append(ids, id)
	}
	return ids
}

func (m *mockCredentialStore) RemoveInstance(id uint64) {
	delete(m.instances, id)
}

var _ application.CredentialStore = (*mockCredentialStore)(nil)

func TestMockCredentialStore_Implements(t *testing.T) {
	store := newMockCredentialStore()

	store.AddInstance(&domain.InstanceCredentials{
		InstanceID: 100,
		APIToken:   "test-token",
		APIURL:     "https://api.green-api.com",
	})

	creds, err := store.GetCredentials(100)
	if err != nil {
		t.Fatalf("GetCredentials: %v", err)
	}
	if creds.APIToken != "test-token" {
		t.Errorf("APIToken: got %q", creds.APIToken)
	}

	ids := store.ListInstances()
	if len(ids) != 1 || ids[0] != 100 {
		t.Errorf("ListInstances: got %v", ids)
	}

	store.RemoveInstance(100)
	_, err = store.GetCredentials(100)
	if err != domain.ErrInstanceNotFound {
		t.Errorf("expected ErrInstanceNotFound, got: %v", err)
	}
}

type mockWhatsAppClient struct{}

func (m *mockWhatsAppClient) CallMethod(_ context.Context, _ uint64, _ string, _ interface{}) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
func (m *mockWhatsAppClient) GetMethod(_ context.Context, _ uint64, _ string) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
func (m *mockWhatsAppClient) RawMethod(_ context.Context, _ uint64, _, _ string, _ interface{}) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
func (m *mockWhatsAppClient) SendMessage(_ context.Context, _ uint64, _ domain.SendMessageRequest) (*domain.SendMessageResponse, error) {
	return &domain.SendMessageResponse{IDMessage: "mock-msg"}, nil
}
func (m *mockWhatsAppClient) SendTemplate(_ context.Context, _ uint64, _ domain.SendTemplateRequest) (*domain.SendMessageResponse, error) {
	return &domain.SendMessageResponse{IDMessage: "mock-tpl"}, nil
}
func (m *mockWhatsAppClient) SendFileByURL(_ context.Context, _ uint64, _ domain.SendFileByURLRequest) (*domain.SendMessageResponse, error) {
	return &domain.SendMessageResponse{IDMessage: "mock-file"}, nil
}
func (m *mockWhatsAppClient) SendFileByUpload(_ context.Context, _ uint64, _, _, _ string, _ []byte) (*domain.SendFileByUploadResponse, error) {
	return &domain.SendFileByUploadResponse{IDMessage: "mock-upload"}, nil
}
func (m *mockWhatsAppClient) GetStateInstance(_ context.Context, _ uint64) (*domain.StateInstanceResponse, error) {
	return &domain.StateInstanceResponse{StateInstance: "authorized"}, nil
}
func (m *mockWhatsAppClient) GetSettings(_ context.Context, _ uint64) (*domain.SettingsResponse, error) {
	return &domain.SettingsResponse{}, nil
}
func (m *mockWhatsAppClient) SetSettings(_ context.Context, _ uint64, _ domain.SetSettingsRequest) (*domain.SetSettingsResponse, error) {
	return &domain.SetSettingsResponse{SaveSettings: true}, nil
}
func (m *mockWhatsAppClient) GetWaSettings(_ context.Context, _ uint64) (*domain.WaSettingsResponse, error) {
	return &domain.WaSettingsResponse{Phone: "79876543210", StateInstance: "authorized"}, nil
}
func (m *mockWhatsAppClient) Reboot(_ context.Context, _ uint64) (json.RawMessage, error) {
	return json.RawMessage(`{"isReboot":true}`), nil
}
func (m *mockWhatsAppClient) CreateTemplate(_ context.Context, _ uint64, _ domain.CreateTemplateRequest) (*domain.TemplateResponse, error) {
	return &domain.TemplateResponse{Template: domain.Template{TemplateID: "tpl-1", Status: domain.TemplateStatusPending}}, nil
}
func (m *mockWhatsAppClient) EditTemplate(_ context.Context, _ uint64, _ domain.EditTemplateRequest) (*domain.TemplateStatusResponse, error) {
	return &domain.TemplateStatusResponse{Status: "success"}, nil
}
func (m *mockWhatsAppClient) GetTemplates(_ context.Context, _ uint64) (*domain.GetTemplatesResponse, error) {
	return &domain.GetTemplatesResponse{}, nil
}
func (m *mockWhatsAppClient) GetTemplateByID(_ context.Context, _ uint64, _ domain.GetTemplateByIDRequest) (*domain.TemplateResponse, error) {
	return &domain.TemplateResponse{}, nil
}
func (m *mockWhatsAppClient) DeleteTemplate(_ context.Context, _ uint64, _ domain.DeleteTemplateRequest) (*domain.TemplateStatusResponse, error) {
	return &domain.TemplateStatusResponse{Status: "success"}, nil
}
func (m *mockWhatsAppClient) DeleteTemplateByID(_ context.Context, _ uint64, _ domain.DeleteTemplateByIDRequest) (*domain.TemplateStatusResponse, error) {
	return &domain.TemplateStatusResponse{Status: "success"}, nil
}
func (m *mockWhatsAppClient) ReceiveNotification(_ context.Context, _ uint64) (*domain.NotificationBody, error) {
	return nil, nil
}
func (m *mockWhatsAppClient) DeleteNotification(_ context.Context, _ uint64, _ int) error {
	return nil
}
func (m *mockWhatsAppClient) GetChatHistory(_ context.Context, _ uint64, _ domain.GetChatHistoryRequest) (json.RawMessage, error) {
	return json.RawMessage(`[]`), nil
}
func (m *mockWhatsAppClient) GetMessage(_ context.Context, _ uint64, _ domain.GetMessageRequest) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
func (m *mockWhatsAppClient) LastIncomingMessages(_ context.Context, _ uint64, _ int) (json.RawMessage, error) {
	return json.RawMessage(`[]`), nil
}
func (m *mockWhatsAppClient) LastOutgoingMessages(_ context.Context, _ uint64, _ int) (json.RawMessage, error) {
	return json.RawMessage(`[]`), nil
}
func (m *mockWhatsAppClient) ShowMessagesQueue(_ context.Context, _ uint64) (json.RawMessage, error) {
	return json.RawMessage(`[]`), nil
}
func (m *mockWhatsAppClient) ClearMessagesQueue(_ context.Context, _ uint64) (json.RawMessage, error) {
	return json.RawMessage(`{"isCleared":true}`), nil
}

var _ application.WhatsAppClient = (*mockWhatsAppClient)(nil)

func TestMockWhatsAppClient_Implements(t *testing.T) {
	client := &mockWhatsAppClient{}
	ctx := context.Background()

	resp, err := client.SendTemplate(ctx, 1, domain.SendTemplateRequest{ChatID: "test@c.us", TemplateID: "tpl-1"})
	if err != nil {
		t.Fatalf("SendTemplate: %v", err)
	}
	if resp.IDMessage != "mock-tpl" {
		t.Errorf("IDMessage: got %q", resp.IDMessage)
	}

	state, err := client.GetStateInstance(ctx, 1)
	if err != nil {
		t.Fatalf("GetStateInstance: %v", err)
	}
	if state.StateInstance != "authorized" {
		t.Errorf("StateInstance: got %q", state.StateInstance)
	}
}

type mockMCPTransport struct{}

func (m *mockMCPTransport) ServeStdio(_ context.Context) error { return nil }

var _ application.MCPTransport = (*mockMCPTransport)(nil)

func TestMockMCPTransport_Implements(t *testing.T) {
	transport := &mockMCPTransport{}
	if err := transport.ServeStdio(context.Background()); err != nil {
		t.Errorf("ServeStdio: unexpected error: %v", err)
	}
}

type mockMetricsProvider struct {
	toolCalls int
	apiCalls  int
}

func (m *mockMetricsProvider) RecordToolCall(_ string, _ time.Time, _ error)   { m.toolCalls++ }
func (m *mockMetricsProvider) RecordAPIRequest(_ string, _ time.Time, _ error) { m.apiCalls++ }
func (m *mockMetricsProvider) SetActiveWebhookConnections(_ float64)           {}
func (m *mockMetricsProvider) RecordRateLimit(_ uint64)                        {}

var _ application.MetricsProvider = (*mockMetricsProvider)(nil)

func TestMockMetricsProvider_Implements(t *testing.T) {
	mp := &mockMetricsProvider{}
	mp.RecordToolCall("test_tool", time.Now(), nil)
	mp.RecordAPIRequest("sendTemplate", time.Now(), nil)
	mp.SetActiveWebhookConnections(5)
	if mp.toolCalls != 1 {
		t.Errorf("toolCalls: got %d, want 1", mp.toolCalls)
	}
	if mp.apiCalls != 1 {
		t.Errorf("apiCalls: got %d, want 1", mp.apiCalls)
	}
}

type mockWebhookBridge struct {
	started bool
	stopped bool
}

func (m *mockWebhookBridge) Start(_ context.Context) error {
	m.started = true
	return nil
}

func (m *mockWebhookBridge) Stop() {
	m.stopped = true
}

var _ application.WebhookBridge = (*mockWebhookBridge)(nil)

func TestMockWebhookBridge_Implements(t *testing.T) {
	bridge := &mockWebhookBridge{}
	if err := bridge.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !bridge.started {
		t.Error("expected started=true")
	}
	bridge.Stop()
	if !bridge.stopped {
		t.Error("expected stopped=true")
	}
}
