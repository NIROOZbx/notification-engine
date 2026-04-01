package provider

import "context"

type mockProvider struct {
	channel string
}

func NewmockProvider(ch string) *mockProvider {
	return &mockProvider{
		channel: ch,
	}

}

func (m *mockProvider) Send(ctx context.Context,msg Message)error{
return nil
}

func(m *mockProvider)Name()string{
return "mock"
}

func (m *mockProvider)Channel()string{
return m.channel
}