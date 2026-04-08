package models


type TriggerPayload struct{
	ExternalUserID string 
	EventType string
	Data map[string]any
	Channels []string
	IdempotencyKey string 
	IsTest bool
}