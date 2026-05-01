package provider

import "context"

type Message struct {
	To      string
	Channel string
	Content map[string]any
}

type Provider interface {
	Send(ctx context.Context, msg Message, config map[string]string) (string,error)
	Channel() string
	Name() string
	RequiredFields() []string
	RequiredContent() []string
}
