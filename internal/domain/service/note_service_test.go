package service

import (
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain/entity"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNoteService_ValidateAndUpdate_Success(t *testing.T) {
	testCases := []struct {
		name        string
		frontMatter map[string]any
		content     string
		expected    bool
	}{
		{
			name: "report",
			frontMatter: map[string]any{
				"title":   "test",
				"source1": nil,
				"author":  "ANkulagin",
				"closed":  false,
			},
			content: `
---
title:  test
source1: 
author: ANkulagin
closed: false
---
anything
`,
			expected: true,
		},
		{
			name: "unreport",
			frontMatter: map[string]any{
				"title":   "test",
				"source1": nil,
				"author":  "ANkulagin",
				"closed":  true,
			},
			content: `
---
title:  test
source1: 
author: ANkulagin
closed: true
---
anything
`,
			expected: false,
		},
		{
			name: "add closed and report",
			frontMatter: map[string]any{
				"title":   "test",
				"source1": nil,
				"author":  "ANkulagin",
			},
			content: `
---
title:  test
source1: 
author: ANkulagin
closed: false
---
anything
`,
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			note := &entity.Note{}
			sut := NewsNoteService()
			note.FrontMatter = tc.frontMatter
			note.Content = tc.content
			actual, err := sut.ValidateAndUpdate(note)
			require.NoError(t, err)
			require.YAMLEq(t, note.Content, tc.content)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestNoteService_ValidateAndUpdate_Error(t *testing.T) {
	testCases := []struct {
		name           string
		content        string
		expectedErrMsg string
	}{
		{
			name: "incorrect front matter format",
			content: `
---
title: test
source1: 
author: ANkulagin
closed: false
`,
			expectedErrMsg: "incorrectly format front matter",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			note := &entity.Note{
				FrontMatter: make(map[string]any),
				Content:     tc.content,
			}
			sut := NewsNoteService()
			_, err := sut.ValidateAndUpdate(note)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedErrMsg)
		})
	}
}
