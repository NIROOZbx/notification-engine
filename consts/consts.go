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
	PlanFree       = "Free"
	PlanPro        = "Pro"
	PlanEnterprise = "Enterprise"
)

const (
	BillingProviderSystem = "system"
	BillingProviderStripe = "stripe"
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
		"resend":true,
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

const (
	FallBackUUID    = "00000000-0000-0000-0000-000000000000"
	OnboardingRoute = "/create-workspace"
	DashboardRoute  = "/dashboard"
)