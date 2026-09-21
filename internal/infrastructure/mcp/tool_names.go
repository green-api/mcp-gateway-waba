package mcp

const (
	ToolPrefix = "waba_"

	ToolConnect    = "waba_connect"
	ToolDisconnect = "waba_disconnect"

	ToolSendMessage      = "waba_send_message"
	ToolSendTemplate     = "waba_send_template"
	ToolSendFile         = "waba_send_file"
	ToolSendFileByUpload = "waba_send_file_by_upload"

	ToolGetState      = "waba_get_state"
	ToolGetSettings   = "waba_get_settings"
	ToolSetSettings   = "waba_set_settings"
	ToolGetWaSettings = "waba_get_wa_settings"
	ToolReboot        = "waba_reboot"

	ToolCreateTemplate     = "waba_create_template"
	ToolEditTemplate       = "waba_edit_template"
	ToolGetTemplates       = "waba_get_templates"
	ToolGetTemplate        = "waba_get_template"
	ToolDeleteTemplate     = "waba_delete_template"
	ToolDeleteTemplateByID = "waba_delete_template_by_id"

	ToolReceiveNotification = "waba_receive_notification"
	ToolDeleteNotification  = "waba_delete_notification"

	ToolGetChatHistory       = "waba_get_chat_history"
	ToolGetMessage           = "waba_get_message"
	ToolLastIncomingMessages = "waba_last_incoming_messages"
	ToolLastOutgoingMessages = "waba_last_outgoing_messages"

	ToolShowMessagesQueue  = "waba_show_messages_queue"
	ToolClearMessagesQueue = "waba_clear_messages_queue"

	ToolServiceMethod = "waba_service_method"

	ToolCreateInstance = "waba_create_instance"
	ToolDeleteInstance = "waba_delete_instance"
	ToolGetInstances   = "waba_get_instances"

	PromptCustomerSupport   = "waba_customer_support"
	PromptTemplateBroadcast = "waba_template_broadcast"

	ResourceTemplatesWidget   = "ui://templates"
	ResourceInstanceState     = "waba://instance/{id}/state"
	ResourceInstanceSettings  = "waba://instance/{id}/settings"
	ResourceInstanceTemplates = "waba://instance/{id}/templates"
)
