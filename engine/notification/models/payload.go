package models

import "time"

type TriggerPayload struct {
	ExternalUserID string
	EventType      string
	Data           map[string]any
	Channels       []string
	IdempotencyKey string
	IsSystem       bool
	IsTest         bool
	ScheduledAt    *time.Time
}
