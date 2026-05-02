package sdk

import "fmt"

// Default API endpoints. These paths are appended to the base URL.
const (
	defaultBaseURL = "http://localhost:8080/api/v1"
	TriggerPath    = "/events/trigger"
	IdentifyPath   = "/identify"
	PreferencePath = "/identify/preferences"
)

// Supported notification channel constants.
const (
	ChannelSMS   = "sms"
	ChannelEmail = "email"
)

// APIError represents an error returned by the Notification Engine API.
// It includes the HTTP status code and the raw response body for debugging.
type APIError struct {
	Code    int
	Message string
}

// Error returns a formatted string representation of the API error,
// satisfying the standard library error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("api error %d: %s", e.Code, e.Message)
}

