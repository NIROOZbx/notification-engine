package delivery

import (
	"fmt"
	"strings"
	"time"

	"github.com/NIROOZbx/notification-engine/pkg/serializer"
)

type sendGridEvent struct {
	Event      string `json:"event"`
	MessageID  string `json:"sg_message_id"`
	Timestamp  int64  `json:"timestamp"`
	Reason     string `json:"reason"`
	BounceType string `json:"type"`
}

func ParseSendGridEvent(body []byte) ([]*DeliveryEvent, error) {
	var events []sendGridEvent

	err := serializer.Unmarshal(body, &events)

	if err != nil {
		return nil, fmt.Errorf("parse sendgrid webhook: %w", err)
	}
	if len(events) == 0 {
        return nil, nil
    }
	var result []*DeliveryEvent
	
    for _, e := range events {
        result = append(result, &DeliveryEvent{
            Provider:          "sendgrid",
            ProviderMessageID: strings.Split(e.MessageID, ".")[0],
            Status:            normalizeSendGridStatus(e.Event),
            Timestamp:         time.Unix(e.Timestamp, 0),
            ErrorMessage:      e.Reason,
        })
    }
    return result, nil

}

func normalizeSendGridStatus(event string) DeliveryStatus {
    switch event {
    case "delivered":
        return StatusDelivered
    case "bounce":
        return StatusBounced
    case "blocked", "dropped":
        return StatusFailed
    case "spamreport":
        return StatusComplained
    default:
        return StatusUnknown
    }
}
