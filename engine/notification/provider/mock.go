package provider

import (
	"context"

	"github.com/rs/zerolog"
)

type mockProvider struct {
	channel string
	log     zerolog.Logger
}

func NewMockProvider(ch string, log zerolog.Logger) *mockProvider {
	return &mockProvider{
		channel: ch,
		log:     log,
	}

}

func (m *mockProvider) Send(ctx context.Context, msg Message, config map[string]string)(string, error) {
	m.log.Info().
		Str("to", msg.To).
		Interface("subject", msg.Content["subject"]).
		Interface("body", msg.Content["body"]).
		Msg("[MOCK] email sent")
	return "",nil
}

func (m *mockProvider) Name() string {
	return "mock"
}

func (m *mockProvider) Channel() string {
	return m.channel
}

func (m *mockProvider) RequiredFields() []string {
	return []string{"subject"}
}
func( m *mockProvider)RequiredContent()[]string{
	return []string{"subject"}
}