package greenapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/green-api/green-api-mcp-gateway-waba/internal/application"
	"github.com/green-api/green-api-mcp-gateway-waba/internal/domain"
	"github.com/green-api/green-api-mcp-gateway-waba/internal/infrastructure/config"
)

type Client struct {
	credentials application.CredentialStore
	metrics     application.MetricsProvider
	config      *config.Config
	httpClient  *http.Client
}

func NewClient(credentials application.CredentialStore, metrics application.MetricsProvider, cfg *config.Config) *Client {
	return &Client{
		credentials: credentials,
		metrics:     metrics,
		config:      cfg,
		httpClient:  &http.Client{Timeout: requestTimeout},
	}
}

func (c *Client) backendBase(instanceID uint64, method string) (string, *domain.InstanceCredentials, error) {
	creds, err := c.credentials.GetCredentials(instanceID)
	if err != nil {
		return "", nil, err
	}
	if c.config != nil && c.config.Routing != nil {
		r := c.config.Routing
		if backendName, ok := r.Methods[method]; ok {
			if bc, ok := r.Backends[backendName]; ok {
				return strings.TrimRight(bc.URL, "/"), creds, nil
			}
		}
		if bc, ok := r.Backends[r.DefaultBackend]; ok {
			return strings.TrimRight(bc.URL, "/"), creds, nil
		}
	}
	return strings.TrimRight(creds.APIURL, "/"), creds, nil
}

func (c *Client) buildURL(instanceID uint64, method string, pathSegments ...string) (string, error) {
	base, creds, err := c.backendBase(instanceID, method)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(pathSegments)+2)
	parts = append(parts, fmt.Sprintf("%s/waInstance%d/%s", base, instanceID, method))
	parts = append(parts, pathSegments...)
	parts = append(parts, creds.APIToken)
	return strings.Join(parts, "/"), nil
}

func (c *Client) httpRequest(ctx context.Context, httpMethod, url string, body interface{}) (json.RawMessage, error) {
	var reqBody io.Reader
	if body != nil && (httpMethod == http.MethodPost || httpMethod == http.MethodPut) {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, httpMethod, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http %s %s: %w", httpMethod, maskURL(url), err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(respData)}
	}

	if len(bytes.TrimSpace(respData)) == 0 {
		return nil, nil //nolint:nilnil
	}

	return json.RawMessage(respData), nil
}

type multipartFile struct {
	Filename string
	Data     []byte
}

func (c *Client) multipartRequest(ctx context.Context, url string, fields map[string]string, files map[string]multipartFile) (json.RawMessage, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for key, value := range fields {
		if err := w.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("write field %s: %w", key, err)
		}
	}
	for fieldName, f := range files {
		fw, err := w.CreateFormFile(fieldName, f.Filename)
		if err != nil {
			return nil, fmt.Errorf("create form file %s: %w", fieldName, err)
		}
		if _, err := fw.Write(f.Data); err != nil {
			return nil, fmt.Errorf("write file data %s: %w", fieldName, err)
		}
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("create multipart request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("multipart request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read multipart response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(respData)}
	}
	return json.RawMessage(respData), nil
}

func toStruct[T any](data json.RawMessage) (*T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &result, nil
}

func httpMethodFor(method string) string {
	switch method {
	case domain.MethodDeleteNotification:
		return http.MethodDelete
	case
		domain.MethodGetStateInstance,
		domain.MethodGetSettings,
		domain.MethodGetWaSettings,
		domain.MethodReboot,
		domain.MethodGetTemplates,
		domain.MethodReceiveNotification,
		domain.MethodShowMessagesQueue,
		domain.MethodClearMessagesQueue,
		domain.MethodLastIncomingMessages,
		domain.MethodLastOutgoingMessages:
		return http.MethodGet
	default:
		return http.MethodPost
	}
}

func (c *Client) call(ctx context.Context, method, verb, url string, body interface{}) (json.RawMessage, error) {
	start := time.Now()
	raw, err := c.httpRequest(ctx, verb, url, body)
	if err != nil {
		wrapped := fmt.Errorf("green api %s: %w", method, err)
		c.metrics.RecordAPIRequest(method, start, wrapped)
		return nil, wrapped
	}
	c.metrics.RecordAPIRequest(method, start, nil)
	return raw, nil
}

func (c *Client) CallMethod(ctx context.Context, instanceID uint64, method string, body interface{}) (json.RawMessage, error) {
	url, err := c.buildURL(instanceID, method)
	if err != nil {
		c.metrics.RecordAPIRequest(method, time.Now(), err)
		return nil, err
	}
	verb := httpMethodFor(method)
	if body != nil && verb == http.MethodGet {
		verb = http.MethodPost
	}
	return c.call(ctx, method, verb, url, body)
}

func (c *Client) GetMethod(ctx context.Context, instanceID uint64, method string) (json.RawMessage, error) {
	url, err := c.buildURL(instanceID, method)
	if err != nil {
		c.metrics.RecordAPIRequest(method, time.Now(), err)
		return nil, err
	}
	return c.call(ctx, method, http.MethodGet, url, nil)
}

func (c *Client) RawMethod(ctx context.Context, instanceID uint64, httpMethod, method string, body interface{}) (json.RawMessage, error) {
	verb := strings.ToUpper(httpMethod)
	if verb != http.MethodGet && verb != http.MethodPost && verb != http.MethodDelete {
		err := fmt.Errorf("unsupported http method %q", verb)
		c.metrics.RecordAPIRequest(method, time.Now(), err)
		return nil, err
	}
	if body != nil && verb != http.MethodPost {
		err := fmt.Errorf("green api %s: a body is only sent with POST, resolved verb is %s", method, verb)
		c.metrics.RecordAPIRequest(method, time.Now(), err)
		return nil, err
	}

	url, err := c.buildURL(instanceID, method)
	if err != nil {
		c.metrics.RecordAPIRequest(method, time.Now(), err)
		return nil, err
	}
	return c.call(ctx, method, verb, url, body)
}

func callTyped[T any](ctx context.Context, c *Client, instanceID uint64, method string, body interface{}) (*T, error) {
	raw, err := c.CallMethod(ctx, instanceID, method, body)
	if err != nil {
		return nil, err
	}
	return toStruct[T](raw)
}

func getTyped[T any](ctx context.Context, c *Client, instanceID uint64, method string) (*T, error) {
	raw, err := c.GetMethod(ctx, instanceID, method)
	if err != nil {
		return nil, err
	}
	return toStruct[T](raw)
}

func (c *Client) SendMessage(ctx context.Context, instanceID uint64, req domain.SendMessageRequest) (*domain.SendMessageResponse, error) {
	return callTyped[domain.SendMessageResponse](ctx, c, instanceID, domain.MethodSendMessage, req)
}

func (c *Client) SendTemplate(ctx context.Context, instanceID uint64, req domain.SendTemplateRequest) (*domain.SendMessageResponse, error) {
	return callTyped[domain.SendMessageResponse](ctx, c, instanceID, domain.MethodSendTemplate, req)
}

func (c *Client) SendFileByURL(ctx context.Context, instanceID uint64, req domain.SendFileByURLRequest) (*domain.SendMessageResponse, error) {
	return callTyped[domain.SendMessageResponse](ctx, c, instanceID, domain.MethodSendFileByURL, req)
}

func (c *Client) SendFileByUpload(ctx context.Context, instanceID uint64, chatID, caption, filename string, fileData []byte) (*domain.SendFileByUploadResponse, error) {
	url, err := c.buildURL(instanceID, domain.MethodSendFileByUpload)
	if err != nil {
		c.metrics.RecordAPIRequest(domain.MethodSendFileByUpload, time.Now(), err)
		return nil, err
	}
	fields := map[string]string{
		"chatId":   chatID,
		"fileName": filename,
	}
	if caption != "" {
		fields["caption"] = caption
	}
	start := time.Now()
	raw, err := c.multipartRequest(ctx, url, fields, map[string]multipartFile{
		"file": {Filename: filename, Data: fileData},
	})
	if err != nil {
		wrapped := fmt.Errorf("green api %s: %w", domain.MethodSendFileByUpload, err)
		c.metrics.RecordAPIRequest(domain.MethodSendFileByUpload, start, wrapped)
		return nil, wrapped
	}
	c.metrics.RecordAPIRequest(domain.MethodSendFileByUpload, start, nil)
	return toStruct[domain.SendFileByUploadResponse](raw)
}

func (c *Client) GetStateInstance(ctx context.Context, instanceID uint64) (*domain.StateInstanceResponse, error) {
	return getTyped[domain.StateInstanceResponse](ctx, c, instanceID, domain.MethodGetStateInstance)
}

func (c *Client) GetSettings(ctx context.Context, instanceID uint64) (*domain.SettingsResponse, error) {
	return getTyped[domain.SettingsResponse](ctx, c, instanceID, domain.MethodGetSettings)
}

func (c *Client) SetSettings(ctx context.Context, instanceID uint64, req domain.SetSettingsRequest) (*domain.SetSettingsResponse, error) {
	return callTyped[domain.SetSettingsResponse](ctx, c, instanceID, domain.MethodSetSettings, req)
}

func (c *Client) GetWaSettings(ctx context.Context, instanceID uint64) (*domain.WaSettingsResponse, error) {
	return getTyped[domain.WaSettingsResponse](ctx, c, instanceID, domain.MethodGetWaSettings)
}

func (c *Client) Reboot(ctx context.Context, instanceID uint64) (json.RawMessage, error) {
	return c.GetMethod(ctx, instanceID, domain.MethodReboot)
}

func (c *Client) CreateTemplate(ctx context.Context, instanceID uint64, req domain.CreateTemplateRequest) (*domain.TemplateResponse, error) {
	return callTyped[domain.TemplateResponse](ctx, c, instanceID, domain.MethodCreateTemplate, req)
}

func (c *Client) EditTemplate(ctx context.Context, instanceID uint64, req domain.EditTemplateRequest) (*domain.TemplateStatusResponse, error) {
	return callTyped[domain.TemplateStatusResponse](ctx, c, instanceID, domain.MethodEditTemplate, req)
}

func (c *Client) GetTemplates(ctx context.Context, instanceID uint64) (*domain.GetTemplatesResponse, error) {
	return getTyped[domain.GetTemplatesResponse](ctx, c, instanceID, domain.MethodGetTemplates)
}

func (c *Client) GetTemplateByID(ctx context.Context, instanceID uint64, req domain.GetTemplateByIDRequest) (*domain.TemplateResponse, error) {
	return callTyped[domain.TemplateResponse](ctx, c, instanceID, domain.MethodGetTemplateByID, req)
}

func (c *Client) DeleteTemplate(ctx context.Context, instanceID uint64, req domain.DeleteTemplateRequest) (*domain.TemplateStatusResponse, error) {
	return callTyped[domain.TemplateStatusResponse](ctx, c, instanceID, domain.MethodDeleteTemplate, req)
}

func (c *Client) DeleteTemplateByID(ctx context.Context, instanceID uint64, req domain.DeleteTemplateByIDRequest) (*domain.TemplateStatusResponse, error) {
	return callTyped[domain.TemplateStatusResponse](ctx, c, instanceID, domain.MethodDeleteTemplateByID, req)
}

func (c *Client) ReceiveNotification(ctx context.Context, instanceID uint64) (*domain.NotificationBody, error) {
	raw, err := c.GetMethod(ctx, instanceID, domain.MethodReceiveNotification)
	if err != nil {
		return nil, err
	}
	if raw == nil || string(raw) == "null" {
		return nil, nil //nolint:nilnil
	}
	return toStruct[domain.NotificationBody](raw)
}

func (c *Client) DeleteNotification(ctx context.Context, instanceID uint64, receiptID int) error {
	url, err := c.buildURL(instanceID, domain.MethodDeleteNotification, strconv.Itoa(receiptID))
	if err != nil {
		c.metrics.RecordAPIRequest(domain.MethodDeleteNotification, time.Now(), err)
		return err
	}
	_, err = c.call(ctx, domain.MethodDeleteNotification, http.MethodDelete, url, nil)
	return err
}

func (c *Client) GetChatHistory(ctx context.Context, instanceID uint64, req domain.GetChatHistoryRequest) (json.RawMessage, error) {
	return c.CallMethod(ctx, instanceID, domain.MethodGetChatHistory, req)
}

func (c *Client) GetMessage(ctx context.Context, instanceID uint64, req domain.GetMessageRequest) (json.RawMessage, error) {
	return c.CallMethod(ctx, instanceID, domain.MethodGetMessage, req)
}

func (c *Client) journal(ctx context.Context, instanceID uint64, method string, minutes int) (json.RawMessage, error) {
	url, err := c.buildURL(instanceID, method)
	if err != nil {
		c.metrics.RecordAPIRequest(method, time.Now(), err)
		return nil, err
	}
	return c.call(ctx, method, http.MethodGet, url+"?minutes="+strconv.Itoa(minutes), nil)
}

func (c *Client) LastIncomingMessages(ctx context.Context, instanceID uint64, minutes int) (json.RawMessage, error) {
	return c.journal(ctx, instanceID, domain.MethodLastIncomingMessages, minutes)
}

func (c *Client) LastOutgoingMessages(ctx context.Context, instanceID uint64, minutes int) (json.RawMessage, error) {
	return c.journal(ctx, instanceID, domain.MethodLastOutgoingMessages, minutes)
}

func (c *Client) ShowMessagesQueue(ctx context.Context, instanceID uint64) (json.RawMessage, error) {
	return c.GetMethod(ctx, instanceID, domain.MethodShowMessagesQueue)
}

func (c *Client) ClearMessagesQueue(ctx context.Context, instanceID uint64) (json.RawMessage, error) {
	return c.GetMethod(ctx, instanceID, domain.MethodClearMessagesQueue)
}
