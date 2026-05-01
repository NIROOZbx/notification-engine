package delivery

import "time"

type DeliveryStatus string

const (
	StatusDelivered  DeliveryStatus = "delivered"
	StatusFailed     DeliveryStatus = "failed"
	StatusBounced    DeliveryStatus = "bounced"
	StatusComplained DeliveryStatus = "complained"
	StatusUnknown    DeliveryStatus = "unknown"
)

type DeliveryEvent struct {
	Provider          string
	ProviderMessageID string
	Status            DeliveryStatus
	Timestamp         time.Time
	ErrorCode         string
	ErrorMessage      string
}