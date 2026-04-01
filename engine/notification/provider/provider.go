package provider

import "context"

type Message struct {
    To      string
    Subject string
    Body    string
}

type Provider interface {
	Send(ctx context.Context,msg Message)error
	Channel() string
	Name() string
}
