package domain

import (
	"errors"
	"gopkg.in/yaml.v2"
	"strings"
)

type Note struct {
	FilePath    string
	FrontMatter map[string]any
	Content     string
}

func (n *Note) HasYaml() bool {
	return strings.HasPrefix(strings.TrimLeft(n.Content, "\t\n\r"), "---")
}

// todo продумать режим где будет просто пропускать сломанный yaml
func (n *Note) LoadFrontMatter() error {
	if !n.HasYaml() {
		return errors.New("front matter not found")
	}
	parts := strings.SplitN(n.Content, "---", 3)
	if len(parts) < 3 {
		return errors.New("incorrectly format front matter")
	}
	if err := yaml.Unmarshal([]byte(parts[1]), &n.FrontMatter); err != nil {
		return err
	}
	return nil
}
