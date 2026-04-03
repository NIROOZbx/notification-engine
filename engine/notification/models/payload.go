package models


type TriggerPayload struct{
	ExternalUserID string 
	EventType string
	Data map[string]any
	IdempotencyKey string 
	IsTest bool
}