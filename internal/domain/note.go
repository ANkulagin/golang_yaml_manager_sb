package domain

import "strings"

type Note struct {
	FilePath    string
	FrontMatter map[string]any
	Content     string
}

func (n *Note) HasYaml() bool {
	return strings.HasPrefix(n.Content, "---")
}
