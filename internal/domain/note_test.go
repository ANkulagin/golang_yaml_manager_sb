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
			expected: true,
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

func TestNote_LoadFrontMatter_Success(t *testing.T) {
	sut := &Note{}

	testCases := []struct {
		name     string
		content  string
		expected map[string]any
	}{
		{
			name: "success",
			content: `
---
title:  wating
source1: 
author: ANkulagin
closed: false
---
`,
			expected: map[string]any{
				"title":   "wating",
				"source1": nil,
				"author":  "ANkulagin",
				"closed":  false,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sut.Content = tc.content

			err := sut.LoadFrontMatter()
			require.NoError(t, err)
			require.Equal(t, tc.expected, sut.FrontMatter)
		})
	}
}

func TestNote_LoadFrontMatter_Error(t *testing.T) {
	sut := &Note{}

	testCases := []struct {
		name           string
		content        string
		expectedErrMsg string
	}{
		{
			name:           "empty content",
			content:        "",
			expectedErrMsg: "front matter not found",
		},
		{
			name:           "--- off the top  of front matter",
			content:        "off the top ---  of front matter",
			expectedErrMsg: "front matter not found",
		},
		{
			name:           "forgot to close",
			content:        "\n---\nclose:false\n forgot to close",
			expectedErrMsg: "incorrectly format front matter",
		},
		{
			name: "incorrectly format front matter",
			content: `
---
title:  wating
source1: 
source2: 
source3: 
source4: 
author: ANkulagin
date: 2025-01-05
description: 
tags: 
closed:false
---
`, //closed: false ✅ closed:false ❌
			expectedErrMsg: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sut.Content = tc.content

			err := sut.LoadFrontMatter()
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedErrMsg)
		})
	}
}

func TestNote_UpdateFrontMatter_Success(t *testing.T) {
	sut := &Note{}
	sut.Content = `
---
title:  test
source1: 
author: ANkulagin
closed: false
---
anything
`

	testCases := []struct {
		name        string
		frontMatter map[string]any
		expected    string
	}{
		{
			name: "success",
			frontMatter: map[string]any{
				"title":   "wating",
				"source1": nil,
				"author":  "ANkulagin",
				"closed":  false,
			},
			expected: `
---
title:  wating
source1:
author: ANkulagin
closed: false
---
anything
`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sut.FrontMatter = tc.frontMatter

			err := sut.UpdateFrontMatter()
			require.NoError(t, err)
			require.YAMLEq(t, tc.expected, sut.Content)
		})
	}
}

func TestNote_UpdateFrontMatter_Error(t *testing.T) {
	sut := &Note{}

	testCases := []struct {
		name           string
		content        string
		expectedErrMsg string
	}{
		{
			name:           "forgot to close",
			content:        "\n---\nclose:false\n forgot to close",
			expectedErrMsg: "incorrectly format front matter",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			sut.Content = tc.content

			err := sut.UpdateFrontMatter()
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedErrMsg)
		})
	}
}
