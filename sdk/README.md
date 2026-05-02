# notification-engine SDK

Go client for the Notification Engine API. Trigger notifications and manage subscriber identities with just an API key.

## Installation

```bash
go get github.com/NIROOZbx/notification-engine/sdk
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/NIROOZbx/notification-engine/sdk"
)

func main() {
	client := sdk.NewClient("your-api-key-here")
	ctx := context.Background()

	// Trigger a notification
	err := client.Notifications.Trigger(ctx, &sdk.TriggerRequest{
		ExternalUserID: "user_123",
		EventType:      "order_confirmed",
		Data: map[string]any{
			"order_id": "ORD-456",
			"total":    99.99,
		},
	})
	if err != nil {
		fmt.Printf("failed: %v\n", err)
		return
	}
	fmt.Println("notification queued")
}
```

## Usage

### Initialize the Client

```go
client := sdk.NewClient("your-api-key")
```

### Trigger a Notification

Sends a notification to a subscriber based on the event type and configured templates.

```go
err := client.Notifications.Trigger(ctx, &sdk.TriggerRequest{
	ExternalUserID: "user_123",
	EventType:      "password_reset",
	Data: map[string]any{
		"reset_link": "https://example.com/reset/abc123",
	},
	Channels:       []string{sdk.ChannelEmail, sdk.ChannelSMS}, // optional: target specific channels
	IdempotencyKey: "unique-key-to-prevent-duplicates",         // optional
	ScheduledAt:    time.Now().Add(10 * time.Minute),           // optional: schedule for later
})
```

### Identify a Subscriber

Registers or updates a subscriber's contact information.

```go
err := client.Contacts.Identify(ctx, &sdk.IdentifyRequest{
	ExternalUserID: "user_123",
	Contacts: []sdk.ContactInfo{
		{Channel: sdk.ChannelEmail, ContactValue: "user@example.com"},
		{Channel: sdk.ChannelSMS, ContactValue: "+1234567890"},
	},
	Metadata: map[string]any{
		"name":  "John Doe",
		"plan":  "pro",
	},
})
```

### Manage Subscriber Preferences

Opt a subscriber in or out of specific notification types.

```go
err := client.Contacts.SetPreference(ctx, &sdk.SetPreferenceRequest{
	ExternalUserID: "user_123",
	Channel:        sdk.ChannelEmail,
	EventType:      "marketing_promo",
	IsEnabled:      false, // unsubscribe from marketing emails
})
```

## Error Handling

The SDK returns typed errors. Use `errors.As` to check for API errors:

```go
import "errors"

err := client.Notifications.Trigger(ctx, req)
if err != nil {
	var apiErr *sdk.APIError
	if errors.As(err, &apiErr) {
		// Handle specific API errors
		switch apiErr.Code {
		case 401:
			fmt.Println("invalid API key")
		case 400:
			fmt.Printf("bad request: %s\n", apiErr.Message)
		case 429:
			fmt.Println("rate limited")
		default:
			fmt.Printf("api error %d: %s\n", apiErr.Code, apiErr.Message)
		}
	} else {
		// Network error or other
		fmt.Printf("request failed: %v\n", err)
	}
}
```

## Channel Constants

Pre-defined constants for supported channels:

- `sdk.ChannelEmail` — `"email"`
- `sdk.ChannelSMS` — `"sms"`

## Requirements

- Go 1.25.1 or later
- A valid API key from your Notification Engine workspace
- Live notification templates configured for the event types you trigger
