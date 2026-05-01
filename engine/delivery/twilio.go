package delivery

import (
    "fmt"
    "net/url"
    "time"
)


func ParseTwilioEvent(body []byte) (*DeliveryEvent, error) {
    values, err := url.ParseQuery(string(body))
    if err != nil {
        return nil, fmt.Errorf("parse twilio webhook: %w", err)
    }

    return &DeliveryEvent{
        Provider:          "twilio",
        ProviderMessageID: values.Get("MessageSid"),
        Status:            normalizeTwilioStatus(values.Get("MessageStatus")),
        Timestamp:         time.Now(),
        ErrorCode:         values.Get("ErrorCode"),
        ErrorMessage:      values.Get("ErrorMessage"),
    }, nil
}

func normalizeTwilioStatus(status string) DeliveryStatus {
    switch status {
    case "delivered":
        return StatusDelivered
    case "failed":
        return StatusFailed
    case "undelivered":
        return StatusBounced
    default:
        return StatusUnknown
    }
}