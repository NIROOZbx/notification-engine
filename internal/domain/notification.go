package domain

import "time"

type UpdateDeliveryStatusInput struct {
    ProviderMessageID string
    DeliveryStatus    string
    Timestamp         time.Time
    ErrorMessage      string
}