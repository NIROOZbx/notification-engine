package sdk

import (
	"context"
	"net/http"
)

// ContactInfo represents a single communication channel for a subscriber.
type ContactInfo struct {
	// Channel is the notification channel (e.g., "email", "sms").
	Channel string `json:"channel"`

	// ContactValue is the address for the channel (email address, phone number, etc.).
	ContactValue string `json:"contact_value"`
}

// IdentifyRequest contains the parameters for identifying or updating a subscriber.
type IdentifyRequest struct {
	// ExternalUserID is the unique identifier for the subscriber in your system.
	ExternalUserID string `json:"external_user_id"`

	// Contacts lists all communication channels and addresses for this subscriber.
	// At least one contact must be provided.
	Contacts []ContactInfo `json:"contacts"`

	// Metadata contains arbitrary key-value pairs to store alongside the subscriber.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Identify registers a new subscriber or updates an existing one's contact information.
//
// Call this before triggering notifications so the engine knows where
// to deliver messages for each channel.
//
// Example:
//
//	err := client.Contacts.Identify(ctx, &sdk.IdentifyRequest{
//		ExternalUserID: "user_123",
//		Contacts: []sdk.ContactInfo{
//			{Channel: sdk.ChannelEmail, ContactValue: "user@example.com"},
//			{Channel: sdk.ChannelSMS, ContactValue: "+1234567890"},
//		},
//	})
func (c *ContactClient) Identify(ctx context.Context, req *IdentifyRequest) error {
	url := c.client.baseURL + IdentifyPath
	return c.client.doRequest(ctx, http.MethodPost, url, req, nil)
}

// SetPreferenceRequest contains the parameters for updating a subscriber's notification preferences.
type SetPreferenceRequest struct {
	// ExternalUserID is the unique identifier of the subscriber.
	ExternalUserID string `json:"external_user_id"`

	// Channel is the notification channel to configure.
	Channel string `json:"channel"`

	// EventType is the specific event type to enable or disable.
	// If empty, applies to all events on the given channel.
	EventType string `json:"event_type"`

	// IsEnabled determines whether notifications are delivered (true) or suppressed (false).
	IsEnabled bool `json:"is_enabled"`
}

// SetPreference updates whether a subscriber receives notifications for a
// specific channel and event type.
//
// Use this to allow subscribers to opt in or out of particular notification types.
//
// Example:
//
//	err := client.Contacts.SetPreference(ctx, &sdk.SetPreferenceRequest{
//		ExternalUserID: "user_123",
//		Channel:        sdk.ChannelEmail,
//		EventType:      "marketing_promo",
//		IsEnabled:      false,
//	})
func (c *ContactClient) SetPreference(ctx context.Context, req *SetPreferenceRequest) error {
	url := c.client.baseURL + PreferencePath
	return c.client.doRequest(ctx, http.MethodPost, url, req, nil)
}
