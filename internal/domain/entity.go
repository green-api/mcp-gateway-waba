package domain

import "encoding/json"

type InstanceCredentials struct {
	InstanceID uint64
	APIToken   string
	APIURL     string
}

type SendMessageRequest struct {
	ChatID      string `json:"chatId"`
	Message     string `json:"message"`
	LinkPreview *bool  `json:"linkPreview,omitempty"`
}

type SendMessageResponse struct {
	IDMessage string `json:"idMessage"`
}

type SendFileByURLRequest struct {
	ChatID   string `json:"chatId"`
	URL      string `json:"urlFile"`
	FileName string `json:"fileName"`
	Caption  string `json:"caption,omitempty"`
}

type SendFileByUploadResponse struct {
	IDMessage string `json:"idMessage"`
	URLFile   string `json:"urlFile,omitempty"`
}

type TemplateMedia struct {
	Link     string `json:"link"`
	Filename string `json:"filename,omitempty"`
}

type TemplateLocation struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Name      string `json:"name,omitempty"`
	Address   string `json:"address,omitempty"`
}

type PostbackText struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

type TemplateCarouselCard struct {
	Link          string         `json:"link"`
	PostbackTexts []PostbackText `json:"postbackTexts,omitempty"`
}

type TemplateMessage struct {
	Type           string                 `json:"type"`
	Image          *TemplateMedia         `json:"image,omitempty"`
	Video          *TemplateMedia         `json:"video,omitempty"`
	Document       *TemplateMedia         `json:"document,omitempty"`
	Location       *TemplateLocation      `json:"location,omitempty"`
	CardHeaderType string                 `json:"cardHeaderType,omitempty"`
	Cards          []TemplateCarouselCard `json:"cards,omitempty"`
}

type SendTemplateRequest struct {
	ChatID        string           `json:"chatId"`
	TemplateID    string           `json:"templateId"`
	Params        []string         `json:"params,omitempty"`
	Message       *TemplateMessage `json:"message,omitempty"`
	PostbackTexts []PostbackText   `json:"postbackTexts,omitempty"`
}

type TemplateButton struct {
	Type        string   `json:"type"`
	Text        string   `json:"text,omitempty"`
	PhoneNumber string   `json:"phone_number,omitempty"`
	URL         string   `json:"url,omitempty"`
	OTPType     string   `json:"otp_type,omitempty"`
	Example     []string `json:"example,omitempty"`
}

type TemplateCard struct {
	HeaderType string           `json:"headerType"`
	MediaURL   string           `json:"mediaUrl"`
	Body       string           `json:"body,omitempty"`
	SampleText string           `json:"sampleText,omitempty"`
	Buttons    []TemplateButton `json:"buttons,omitempty"`
}

type CreateTemplateRequest struct {
	ElementName                 string `json:"elementName"`
	LanguageCode                string `json:"languageCode"`
	Category                    string `json:"category"`
	TemplateType                string `json:"templateType"`
	Vertical                    string `json:"vertical"`
	Content                     string `json:"content,omitempty"`
	Example                     string `json:"example,omitempty"`
	Header                      string `json:"header,omitempty"`
	ExampleHeader               string `json:"exampleHeader,omitempty"`
	Footer                      string `json:"footer,omitempty"`
	Buttons                     string `json:"buttons,omitempty"`
	Cards                       string `json:"cards,omitempty"`
	MediaURL                    string `json:"mediaUrl,omitempty"`
	EnableSample                bool   `json:"enableSample"`
	AllowTemplateCategoryChange bool   `json:"allowTemplateCategoryChange,omitempty"`
	AddSecurityRecommendation   bool   `json:"addSecurityRecommendation,omitempty"`
	CodeExpirationMinutes       int    `json:"codeExpirationMinutes,omitempty"`
}

type EditTemplateParams struct {
	Content                     string `json:"content,omitempty"`
	TemplateType                string `json:"templateType,omitempty"`
	Example                     string `json:"example,omitempty"`
	EnableSample                *bool  `json:"enableSample,omitempty"`
	Header                      string `json:"header,omitempty"`
	ExampleHeader               string `json:"exampleHeader,omitempty"`
	Footer                      string `json:"footer,omitempty"`
	Buttons                     string `json:"buttons,omitempty"`
	ExampleMedia                string `json:"exampleMedia,omitempty"`
	MediaID                     string `json:"mediaId,omitempty"`
	MediaURL                    string `json:"mediaUrl,omitempty"`
	Category                    string `json:"category,omitempty"`
	AllowTemplateCategoryChange *bool  `json:"allowTemplateCategoryChange,omitempty"`
}

type EditTemplateRequest struct {
	TemplateID     string             `json:"templateId"`
	TemplateParams EditTemplateParams `json:"templateParams"`
}

type DeleteTemplateRequest struct {
	ElementName string `json:"elementName"`
}

type DeleteTemplateByIDRequest struct {
	ElementName string `json:"elementName"`
	TemplateID  string `json:"templateId"`
}

type GetTemplateByIDRequest struct {
	TemplateID string `json:"templateId"`
}

type Template struct {
	TemplateID      string `json:"templateId,omitempty"`
	ID              string `json:"id,omitempty"`
	ElementName     string `json:"elementName"`
	Status          string `json:"status"`
	Category        string `json:"category"`
	TemplateType    string `json:"templateType"`
	LanguageCode    string `json:"languageCode"`
	Data            string `json:"data,omitempty"`
	ContainerMeta   string `json:"containerMeta,omitempty"`
	Meta            string `json:"meta,omitempty"`
	Vertical        string `json:"vertical,omitempty"`
	ButtonSupported string `json:"buttonSupported,omitempty"`
	Quality         string `json:"quality,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	Stage           string `json:"stage,omitempty"`
	CreatedOn       int64  `json:"createdOn,omitempty"`
	ModifiedOn      int64  `json:"modifiedOn,omitempty"`
}

type GetTemplatesResponse struct {
	Templates []Template `json:"templates"`
}

type TemplateResponse struct {
	Template Template `json:"template"`
}

type TemplateStatusResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type StateInstanceResponse struct {
	StateInstance string `json:"stateInstance"`
}

type SettingsResponse struct {
	Wid                       string `json:"wid,omitempty"`
	WebhookURL                string `json:"webhookUrl"`
	WebhookURLToken           string `json:"webhookUrlToken,omitempty"`
	OutgoingWebhook           string `json:"outgoingWebhook,omitempty"`
	OutgoingAPIMessageWebhook string `json:"outgoingAPIMessageWebhook,omitempty"`
	IncomingWebhook           string `json:"incomingWebhook,omitempty"`
}

type SetSettingsRequest struct {
	WebhookURL                *string `json:"webhookUrl,omitempty"`
	WebhookURLToken           *string `json:"webhookUrlToken,omitempty"`
	OutgoingWebhook           string  `json:"outgoingWebhook,omitempty"`
	OutgoingAPIMessageWebhook string  `json:"outgoingAPIMessageWebhook,omitempty"`
	IncomingWebhook           string  `json:"incomingWebhook,omitempty"`
}

type SetSettingsResponse struct {
	SaveSettings bool `json:"saveSettings"`
}

type WaSettingsResponse struct {
	Avatar        string `json:"avatar"`
	Phone         string `json:"phone"`
	StateInstance string `json:"stateInstance"`
	DeviceID      string `json:"deviceId"`
}

type NotificationBody struct {
	ReceiptID int             `json:"receiptId"`
	Body      json.RawMessage `json:"body"`
}

type GetChatHistoryRequest struct {
	ChatID string `json:"chatId"`
	Count  int    `json:"count,omitempty"`
}

type GetMessageRequest struct {
	ChatID    string `json:"chatId"`
	IDMessage string `json:"idMessage"`
}
