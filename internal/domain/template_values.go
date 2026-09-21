package domain

const (
	TemplateCategoryAuthentication = "AUTHENTICATION"
	TemplateCategoryMarketing      = "MARKETING"
	TemplateCategoryUtility        = "UTILITY"

	TemplateTypeText     = "TEXT"
	TemplateTypeImage    = "IMAGE"
	TemplateTypeVideo    = "VIDEO"
	TemplateTypeDocument = "DOCUMENT"
	TemplateTypeCarousel = "CAROUSEL"

	TemplateStatusPending  = "PENDING"
	TemplateStatusApproved = "APPROVED"
	TemplateStatusRejected = "REJECTED"
	TemplateStatusFailed   = "FAILED"
	TemplateStatusPaused   = "PAUSED"

	TemplateButtonQuickReply  = "QUICK_REPLY"
	TemplateButtonURL         = "URL"
	TemplateButtonPhoneNumber = "PHONE_NUMBER"
	TemplateButtonOTP         = "OTP"

	TemplateMediaImage    = "image"
	TemplateMediaVideo    = "video"
	TemplateMediaDocument = "document"
	TemplateMediaLocation = "location"
	TemplateMediaCarousel = "carousel"

	SettingEnabled  = "yes"
	SettingDisabled = "no"
)

var (
	TemplateCategories = []string{TemplateCategoryAuthentication, TemplateCategoryMarketing, TemplateCategoryUtility}
	TemplateTypes      = []string{TemplateTypeText, TemplateTypeImage, TemplateTypeVideo, TemplateTypeDocument, TemplateTypeCarousel}
	TemplateStatuses   = []string{TemplateStatusPending, TemplateStatusApproved, TemplateStatusRejected, TemplateStatusFailed, TemplateStatusPaused}
	TemplateMediaTypes = []string{TemplateMediaImage, TemplateMediaVideo, TemplateMediaDocument}
)
