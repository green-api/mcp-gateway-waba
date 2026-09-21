package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/green-api/green-api-mcp-gateway-waba/internal/domain"
)

type CredentialStore interface {
	AddInstance(creds *domain.InstanceCredentials)
	GetCredentials(instanceID uint64) (*domain.InstanceCredentials, error)
	ListInstances() []uint64
	RemoveInstance(instanceID uint64)
}

type WhatsAppClient interface {
	CallMethod(ctx context.Context, instanceID uint64, method string, body interface{}) (json.RawMessage, error)
	GetMethod(ctx context.Context, instanceID uint64, method string) (json.RawMessage, error)

	RawMethod(ctx context.Context, instanceID uint64, httpMethod, method string, body interface{}) (json.RawMessage, error)

	SendMessage(ctx context.Context, instanceID uint64, req domain.SendMessageRequest) (*domain.SendMessageResponse, error)
	SendTemplate(ctx context.Context, instanceID uint64, req domain.SendTemplateRequest) (*domain.SendMessageResponse, error)
	SendFileByURL(ctx context.Context, instanceID uint64, req domain.SendFileByURLRequest) (*domain.SendMessageResponse, error)
	SendFileByUpload(ctx context.Context, instanceID uint64, chatID, caption, filename string, fileData []byte) (*domain.SendFileByUploadResponse, error)

	GetStateInstance(ctx context.Context, instanceID uint64) (*domain.StateInstanceResponse, error)
	GetSettings(ctx context.Context, instanceID uint64) (*domain.SettingsResponse, error)
	SetSettings(ctx context.Context, instanceID uint64, req domain.SetSettingsRequest) (*domain.SetSettingsResponse, error)
	GetWaSettings(ctx context.Context, instanceID uint64) (*domain.WaSettingsResponse, error)
	Reboot(ctx context.Context, instanceID uint64) (json.RawMessage, error)

	CreateTemplate(ctx context.Context, instanceID uint64, req domain.CreateTemplateRequest) (*domain.TemplateResponse, error)
	EditTemplate(ctx context.Context, instanceID uint64, req domain.EditTemplateRequest) (*domain.TemplateStatusResponse, error)
	GetTemplates(ctx context.Context, instanceID uint64) (*domain.GetTemplatesResponse, error)
	GetTemplateByID(ctx context.Context, instanceID uint64, req domain.GetTemplateByIDRequest) (*domain.TemplateResponse, error)
	DeleteTemplate(ctx context.Context, instanceID uint64, req domain.DeleteTemplateRequest) (*domain.TemplateStatusResponse, error)
	DeleteTemplateByID(ctx context.Context, instanceID uint64, req domain.DeleteTemplateByIDRequest) (*domain.TemplateStatusResponse, error)

	ReceiveNotification(ctx context.Context, instanceID uint64) (*domain.NotificationBody, error)
	DeleteNotification(ctx context.Context, instanceID uint64, receiptID int) error

	GetChatHistory(ctx context.Context, instanceID uint64, req domain.GetChatHistoryRequest) (json.RawMessage, error)
	GetMessage(ctx context.Context, instanceID uint64, req domain.GetMessageRequest) (json.RawMessage, error)
	LastIncomingMessages(ctx context.Context, instanceID uint64, minutes int) (json.RawMessage, error)
	LastOutgoingMessages(ctx context.Context, instanceID uint64, minutes int) (json.RawMessage, error)

	ShowMessagesQueue(ctx context.Context, instanceID uint64) (json.RawMessage, error)
	ClearMessagesQueue(ctx context.Context, instanceID uint64) (json.RawMessage, error)
}

type MCPTransport interface {
	ServeStdio(ctx context.Context) error
}

type MetricsProvider interface {
	RecordToolCall(tool string, start time.Time, err error)
	RecordAPIRequest(method string, start time.Time, err error)
	SetActiveWebhookConnections(count float64)
	RecordRateLimit(instanceID uint64)
}

type WebhookBridge interface {
	Start(ctx context.Context) error
	Stop()
}
