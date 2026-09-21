package domain

const (
	MethodGetSettings      = "getSettings"
	MethodSetSettings      = "setSettings"
	MethodGetStateInstance = "getStateInstance"
	MethodReboot           = "reboot"
	MethodGetWaSettings    = "getWaSettings"

	MethodSendMessage      = "sendMessage"
	MethodSendTemplate     = "sendTemplate"
	MethodSendFileByURL    = "sendFileByUrl"
	MethodSendFileByUpload = "sendFileByUpload"

	MethodCreateTemplate     = "createTemplate"
	MethodEditTemplate       = "editTemplate"
	MethodGetTemplates       = "getTemplates"
	MethodGetTemplateByID    = "getTemplateById"
	MethodDeleteTemplate     = "deleteTemplate"
	MethodDeleteTemplateByID = "deleteTemplateById"

	MethodReceiveNotification = "receiveNotification"
	MethodDeleteNotification  = "deleteNotification"

	MethodGetChatHistory       = "getChatHistory"
	MethodGetMessage           = "getMessage"
	MethodLastIncomingMessages = "lastIncomingMessages"
	MethodLastOutgoingMessages = "lastOutgoingMessages"

	MethodShowMessagesQueue  = "showMessagesQueue"
	MethodClearMessagesQueue = "clearMessagesQueue"
)
