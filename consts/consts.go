package consts

import "time"

const (
	UID         = "uid"
	WID         = "wid"
	ENVID       = "envID"
	KEYID       = "keyID"
	ISTEST      = "isTest"
	Role        = "role"
	MaxAttempts = 5
	Interval    = 30 * time.Second
)

const (
	ChannelEmail    = "email"
	ChannelSMS      = "sms"
	ChannelPush     = "push"
	ChannelSlack    = "slack"
	ChannelWhatsApp = "whatsapp"
	ChannelWebhook  = "webhook"
	ChannelInApp    = "in_app"
)

var ValidChannels = map[string]bool{
	ChannelEmail:    true,
	ChannelSMS:      true,
	ChannelPush:     true,
	ChannelSlack:    true,
	ChannelWhatsApp: true,
	ChannelWebhook:  true,
	ChannelInApp:    true,
}
var ValidProvidersByChannel = map[string]map[string]bool{
	ChannelEmail: {
		"sendgrid": true,
		"mailgun":  true,
		"ses":      true,
	},
	ChannelSMS: {
		"twilio": true,
		"vonage": true,
	},
	ChannelPush: {
		"fcm":  true,
		"apns": true,
	},
}
