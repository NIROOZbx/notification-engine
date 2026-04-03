package models

type NotificationEvent struct {
	NotificationLogID string
	WorkspaceID       string
	EnvironmentID     string
	Channel           string
	Data map[string]any
	AttemptNumber     int
	Recipient string
}
