package tools

type NotificationTool struct{}

func NewNotificationTool() *NotificationTool {
	return &NotificationTool{}
}

func (t *NotificationTool) SendNotification(phoneNumber, message string) string {
	return `{"success":true,"phone":"` + phoneNumber + `","messageSent":true,"smsId":"SMS-` + phoneNumber + `"}`
}
