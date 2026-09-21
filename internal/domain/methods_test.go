package domain

import "testing"

var allMethods = map[string]string{
	"GetSettings":          MethodGetSettings,
	"SetSettings":          MethodSetSettings,
	"GetStateInstance":     MethodGetStateInstance,
	"Reboot":               MethodReboot,
	"GetWaSettings":        MethodGetWaSettings,
	"SendMessage":          MethodSendMessage,
	"SendTemplate":         MethodSendTemplate,
	"SendFileByURL":        MethodSendFileByURL,
	"SendFileByUpload":     MethodSendFileByUpload,
	"CreateTemplate":       MethodCreateTemplate,
	"EditTemplate":         MethodEditTemplate,
	"GetTemplates":         MethodGetTemplates,
	"GetTemplateByID":      MethodGetTemplateByID,
	"DeleteTemplate":       MethodDeleteTemplate,
	"DeleteTemplateByID":   MethodDeleteTemplateByID,
	"ReceiveNotification":  MethodReceiveNotification,
	"DeleteNotification":   MethodDeleteNotification,
	"GetChatHistory":       MethodGetChatHistory,
	"GetMessage":           MethodGetMessage,
	"LastIncomingMessages": MethodLastIncomingMessages,
	"LastOutgoingMessages": MethodLastOutgoingMessages,
	"ShowMessagesQueue":    MethodShowMessagesQueue,
	"ClearMessagesQueue":   MethodClearMessagesQueue,
}

func TestMethodConstants_Values(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"SendMessage", MethodSendMessage, "sendMessage"},
		{"SendTemplate", MethodSendTemplate, "sendTemplate"},
		{"SendFileByURL", MethodSendFileByURL, "sendFileByUrl"},
		{"SendFileByUpload", MethodSendFileByUpload, "sendFileByUpload"},
		{"GetStateInstance", MethodGetStateInstance, "getStateInstance"},
		{"GetSettings", MethodGetSettings, "getSettings"},
		{"SetSettings", MethodSetSettings, "setSettings"},
		{"GetWaSettings", MethodGetWaSettings, "getWaSettings"},
		{"Reboot", MethodReboot, "reboot"},
		{"CreateTemplate", MethodCreateTemplate, "createTemplate"},
		{"EditTemplate", MethodEditTemplate, "editTemplate"},
		{"GetTemplates", MethodGetTemplates, "getTemplates"},
		{"GetTemplateByID", MethodGetTemplateByID, "getTemplateById"},
		{"DeleteTemplate", MethodDeleteTemplate, "deleteTemplate"},
		{"DeleteTemplateByID", MethodDeleteTemplateByID, "deleteTemplateById"},
		{"ReceiveNotification", MethodReceiveNotification, "receiveNotification"},
		{"DeleteNotification", MethodDeleteNotification, "deleteNotification"},
		{"GetChatHistory", MethodGetChatHistory, "getChatHistory"},
		{"GetMessage", MethodGetMessage, "getMessage"},
		{"LastIncomingMessages", MethodLastIncomingMessages, "lastIncomingMessages"},
		{"LastOutgoingMessages", MethodLastOutgoingMessages, "lastOutgoingMessages"},
		{"ShowMessagesQueue", MethodShowMessagesQueue, "showMessagesQueue"},
		{"ClearMessagesQueue", MethodClearMessagesQueue, "clearMessagesQueue"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.want)
			}
		})
	}
}

func TestMethodConstants_NotEmptyAndUnique(t *testing.T) {
	seen := make(map[string]string, len(allMethods))
	for name, m := range allMethods {
		if m == "" {
			t.Errorf("%s: method constant should not be empty", name)
		}
		if other, dup := seen[m]; dup {
			t.Errorf("%s and %s share the value %q", name, other, m)
		}
		seen[m] = name
	}
	if len(allMethods) != 23 {
		t.Errorf("expected 23 WABA methods, got %d", len(allMethods))
	}
}

func TestTemplateValueLists(t *testing.T) {
	if len(TemplateCategories) != 3 || len(TemplateTypes) != 5 || len(TemplateStatuses) != 5 || len(TemplateMediaTypes) != 3 {
		t.Errorf("unexpected list sizes: %d %d %d %d", len(TemplateCategories), len(TemplateTypes), len(TemplateStatuses), len(TemplateMediaTypes))
	}
	if SettingEnabled != "yes" || SettingDisabled != "no" {
		t.Errorf("setting flags: %q %q", SettingEnabled, SettingDisabled)
	}
}
