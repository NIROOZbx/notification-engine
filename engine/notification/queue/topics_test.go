package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopicByChannel(t *testing.T) {
	tests := []struct {
		name          string
		channel       string
		expectedTopic string
		expectError   bool
	}{
		{
			name:          "Valid: Email",
			channel:       "email",
			expectedTopic: TopicEmail,
			expectError:   false,
		},
		{
			name:          "Valid: SMS",
			channel:       "sms",
			expectedTopic: TopicSMS,
			expectError:   false,
		},
		{
			name:          "Valid: In-App",
			channel:       "in_app",
			expectedTopic: TopicInApp,
			expectError:   false,
		},
		{
			name:          "Invalid: Pigeon",
			channel:       "pigeon",
			expectedTopic: "",
			expectError:   true,
		},
		{
			name:          "Invalid: Empty string",
			channel:       "",
			expectedTopic: "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic, err := TopicByChannel(tt.channel)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, topic)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTopic, topic)
			}
		})
	}
}
