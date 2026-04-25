package queue

import "fmt"

const (
	TopicEmail    = "notifications.email"
	TopicSMS      = "notifications.sms"
	TopicPush     = "notifications.push"
	TopicSlack    = "notifications.slack"
	TopicWhatsApp = "notifications.whatsapp"
	TopicWebhook  = "notifications.webhook"
	TopicInApp    = "notifications.in_app"
	TopicRetry    = "notifications.retry"
	TopicDLQ      = "notifications.dlq"
	TopicSystem ="system.notifications"
)

func TopicByChannel(channel string) (string, error) {
	switch channel {
	case "email":
		return TopicEmail, nil
	case "sms":
		return TopicSMS, nil
	case "push":
		return TopicPush, nil
	case "slack":
		return TopicSlack, nil
	case "whatsapp":
		return TopicWhatsApp, nil
	case "webhook":
		return TopicWebhook, nil
	case "in_app":
		return TopicInApp, nil
	default:
		return "", fmt.Errorf("unsupported or unknown channel: %s", channel)
	}
}