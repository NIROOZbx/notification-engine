package sdk

import (
	"context"
	"net/http"
	"time"
)


// TriggerRequest contains the parameters for triggering a notification event.
// ExternalUserID and EventType are required; all other fields are optional.
type TriggerRequest struct {
	// ExternalUserID is the unique identifier of the subscriber to notify.
	ExternalUserID string `json:"external_user_id"`

	// EventType identifies the notification template to use (e.g., "order_confirmed").
	EventType string `json:"event_type"`

	// Data contains template variables for rendering the notification content.
	Data map[string]any `json:"data"`

	// Channels limits delivery to specific channels. If empty, all active
	// channels configured for the template are used.
	Channels []string `json:"channels"`

	// IdempotencyKey prevents duplicate notifications when retried.
	// Use a stable unique key (e.g., "order_123:email") for safe retries.
	IdempotencyKey string `json:"idempotency_key"`

	// ScheduledAt defers delivery until the specified time.
	// If nil or in the past, the notification is sent immediately.
	ScheduledAt *time.Time `json:"scheduled_at"`
}

// Trigger sends a notification event to the specified subscriber.
//
// The notification is queued asynchronously and delivered via the
// channels configured for the event type template.
//
// Example:
//
//	err := client.Notifications.Trigger(ctx, &sdk.TriggerRequest{
//		ExternalUserID: "user_123",
//		EventType:      "order_confirmed",
//		Data: map[string]any{
//			"order_id": "ORD-456",
//			"total":    99.99,
//		},
//	})
func (c *NotificationClient) Trigger(ctx context.Context, input *TriggerRequest) error {
	url := c.client.baseURL + TriggerPath
	return c.client.doRequest(ctx, http.MethodPost, url, input, nil)
}