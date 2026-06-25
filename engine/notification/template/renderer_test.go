package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderer_Render(t *testing.T) {
	renderer := NewRenderer()

	tests := []struct {
		name          string
		templateStr   string
		data          map[string]any
		expectedStr   string
		expectedError string
	}{
		{
			name:        "Success: Single variable",
			templateStr: "Hello {{name}}!",
			data:        map[string]any{"name": "Alice"},
			expectedStr: "Hello Alice!",
		},
		{
			name:        "Success: Multiple variables",
			templateStr: "{{greeting}} {{name}}, your order {{order_id}} is ready.",
			data: map[string]any{
				"greeting": "Hi",
				"name":     "Bob",
				"order_id": 12345,
			},
			expectedStr: "Hi Bob, your order 12345 is ready.",
		},
		{
			name:        "Success: Repeated variable",
			templateStr: "{{name}} said {{name}} is coming.",
			data:        map[string]any{"name": "Charlie"},
			expectedStr: "Charlie said Charlie is coming.",
		},
		{
			name:        "Success: No variables in template",
			templateStr: "Just static text here.",
			data:        map[string]any{"ignored": "data"},
			expectedStr: "Just static text here.",
		},
		{
			name:          "Fail: Missing variable",
			templateStr:   "Hello {{name}}, you have {{count}} items.",
			data:          map[string]any{"name": "Dave"},
			expectedError: "missing template variable: \"count\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderer.Render(tt.templateStr, tt.data)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStr, result)
			}
		})
	}
}
