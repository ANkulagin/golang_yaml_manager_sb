package entity

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

func (n *Note) CheckHasYaml() bool {
	return strings.HasPrefix(strings.TrimLeft(n.Content, "\t\n\r"), "---")
}

// todo продумать режим где будет просто пропускать сломанный yaml
func (n *Note) FillFrontMatter() error {
	if !n.CheckHasYaml() {
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

func (n *Note) UpdateFrontMatter() error {
	data, err := yaml.Marshal(n.FrontMatter)
	if err != nil {
		return err
	}
	parts := strings.SplitN(n.Content, "---", 3)
	if len(parts) < 3 {
		return errors.New("incorrectly format front matter")
	}
	n.Content = "---\n" + string(data) + "---" + parts[2]
	return nil
}
