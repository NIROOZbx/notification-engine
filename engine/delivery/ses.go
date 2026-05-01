package delivery

import (
	"fmt"
	"time"

	"github.com/NIROOZbx/notification-engine/pkg/serializer"
)

type snsEnvelope struct {
	Type    string `json:"Type"`
	Message string `json:"Message"`
}

type sesEvent struct {
	EventType string `json:"eventType"`
	Mail      struct {
		MessageID string `json:"messageId"`
		Timestamp string `json:"timestamp"`
	} `json:"mail"`
	Bounce *struct {
		BounceType    string `json:"bounceType"`
		BounceSubType string `json:"bounceSubType"`
	} `json:"bounce"`
	Complaint *struct {
		FeedbackType string `json:"feedbackType"`
	} `json:"complaint"`
	Failure *struct {
		ErrorMessage string `json:"errorMessage"`
	} `json:"failure"`
}

func ParseSesEvent(body []byte) (*DeliveryEvent, error) {
	var sns snsEnvelope

	err:=serializer.Unmarshal(body,&sns)
	if err!=nil{
		return nil, fmt.Errorf("parse sns envelope: %w", err)
	}
	if sns.Type != "Notification" {
		return nil, nil 
	}

	var ses  sesEvent

	if err:=serializer.Unmarshal([]byte(sns.Message),&ses);err!=nil{
		return nil, fmt.Errorf("parse ses event: %w", err)
	}
	    ts, err := time.Parse(time.RFC3339, ses.Mail.Timestamp)
		if err!=nil{
			return nil,fmt.Errorf("parsing time error:%w",err)
		}
		event := &DeliveryEvent{
        Provider:          "ses",
        ProviderMessageID: ses.Mail.MessageID,
        Status:            normalizeSESStatus(ses.EventType),
        Timestamp:         ts,
    }
	if ses.Bounce != nil {
        event.ErrorMessage = fmt.Sprintf("%s %s", ses.Bounce.BounceType, ses.Bounce.BounceSubType)
    }
    if ses.Failure != nil {
        event.ErrorMessage = ses.Failure.ErrorMessage
    }

	return event,nil



}
func normalizeSESStatus(eventType string) DeliveryStatus {
    switch eventType {
    case "Delivery":
        return StatusDelivered
    case "Bounce":
        return StatusBounced
    case "Complaint":
        return StatusComplained
    case "Send", "Reject", "RenderingFailure":
        return StatusFailed
    default:
        return StatusUnknown
    }
}