package provider

import "context"

type mockProvider struct {
	channel string
}

func NewMockProvider(ch string) *mockProvider {
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

func(m *mockProvider)RequiredFields()[]string{
	return []string{"subject"}
}