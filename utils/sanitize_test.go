package utils_test

import (
	"testing"

	"github.com/teknologi-umum/captcha/utils"
)

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Sanitize simple string",
			input:    "Hello, world!",
			expected: "Hello, world!",
		},
		{
			name:     "Sanitize string with greater than",
			input:    "This is > than that",
			expected: "This is &gt; than that",
		},
		{
			name:     "Sanitize string with less than",
			input:    "This is < than that",
			expected: "This is &lt; than that",
		},
		{
			name:     "Sanitize string with both greater than and less than",
			input:    "This is > than that, but < than the other",
			expected: "This is &gt; than that, but &lt; than the other",
		},
		{
			name:     "Sanitize empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function to be tested
			actual := utils.SanitizeInput(tc.input)

			// Assert the expected result
			if actual != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, actual)
			}
		})
	}
}
