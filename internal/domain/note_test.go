package domain

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNote_HasYaml(t *testing.T) {
	sut := &Note{}

	testCases := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "empty content",
			content:  "",
			expected: false,
		},
		{
			name:     "content without yaml",
			content:  "Hello, world!",
			expected: false,
		},
		{
			name:     "content with yaml",
			content:  "---\nHello, world!\n---",
			expected: true,
		},
		{
			name:     "yaml is a hat but does not go at the beginning of the file, but in the middle",
			content:  "content test \n---\nHello, world!\n---",
			expected: false,
		},
		{
			name:     "empty line at the beginning of the file",
			content:  "\n\n---\nHello, world!\n---",
			expected: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sut.Content = tc.content

			result := sut.HasYaml()

			require.Equal(t, tc.expected, result)
		})
	}
}
