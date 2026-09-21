package infrastructure

import (
	"context"
	"encoding/json"

	"github.com/green-api/green-api-mcp-gateway-waba/internal/domain"
)

type MockWhatsAppClient struct {
	GetStateInstanceFunc func(ctx context.Context, instanceID uint64) (*domain.StateInstanceResponse, error)
}

func (m *MockWhatsAppClient) GetStateInstance(ctx context.Context, instanceID uint64) (*domain.StateInstanceResponse, error) {
	if m.GetStateInstanceFunc != nil {
		return m.GetStateInstanceFunc(ctx, instanceID)
	}
	return &domain.StateInstanceResponse{StateInstance: "authorized"}, nil
}

func (m *MockWhatsAppClient) CallMethod(_ context.Context, _ uint64, _ string, _ interface{}) (json.RawMessage, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) GetMethod(_ context.Context, _ uint64, _ string) (json.RawMessage, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) RawMethod(_ context.Context, _ uint64, _, _ string, _ interface{}) (json.RawMessage, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) SendMessage(_ context.Context, _ uint64, _ domain.SendMessageRequest) (*domain.SendMessageResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) SendTemplate(_ context.Context, _ uint64, _ domain.SendTemplateRequest) (*domain.SendMessageResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) SendFileByURL(_ context.Context, _ uint64, _ domain.SendFileByURLRequest) (*domain.SendMessageResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) SendFileByUpload(_ context.Context, _ uint64, _, _, _ string, _ []byte) (*domain.SendFileByUploadResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) GetSettings(_ context.Context, _ uint64) (*domain.SettingsResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) SetSettings(_ context.Context, _ uint64, _ domain.SetSettingsRequest) (*domain.SetSettingsResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) GetWaSettings(_ context.Context, _ uint64) (*domain.WaSettingsResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) Reboot(_ context.Context, _ uint64) (json.RawMessage, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) CreateTemplate(_ context.Context, _ uint64, _ domain.CreateTemplateRequest) (*domain.TemplateResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) EditTemplate(_ context.Context, _ uint64, _ domain.EditTemplateRequest) (*domain.TemplateStatusResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) GetTemplates(_ context.Context, _ uint64) (*domain.GetTemplatesResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) GetTemplateByID(_ context.Context, _ uint64, _ domain.GetTemplateByIDRequest) (*domain.TemplateResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) DeleteTemplate(_ context.Context, _ uint64, _ domain.DeleteTemplateRequest) (*domain.TemplateStatusResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) DeleteTemplateByID(_ context.Context, _ uint64, _ domain.DeleteTemplateByIDRequest) (*domain.TemplateStatusResponse, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) ReceiveNotification(_ context.Context, _ uint64) (*domain.NotificationBody, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) DeleteNotification(_ context.Context, _ uint64, _ int) error {
	return nil
}
func (m *MockWhatsAppClient) GetChatHistory(_ context.Context, _ uint64, _ domain.GetChatHistoryRequest) (json.RawMessage, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) GetMessage(_ context.Context, _ uint64, _ domain.GetMessageRequest) (json.RawMessage, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) LastIncomingMessages(_ context.Context, _ uint64, _ int) (json.RawMessage, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) LastOutgoingMessages(_ context.Context, _ uint64, _ int) (json.RawMessage, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) ShowMessagesQueue(_ context.Context, _ uint64) (json.RawMessage, error) {
	return nil, nil
}
func (m *MockWhatsAppClient) ClearMessagesQueue(_ context.Context, _ uint64) (json.RawMessage, error) {
	return nil, nil
}
