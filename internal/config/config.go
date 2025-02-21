package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	SrcDir           string   `yaml:"src_dir"`
	TemplateDir      string   `yaml:"template_dir"`
	ReportFile       string   `yaml:"report_file"`
	LogLevel         string   `yaml:"log_level"`
	ConcurrencyLimit int      `yaml:"concurrency_limit"`
	SkipPatterns     []string `yaml:"skip_patterns"`
}

func LoadConfig(configPath string) (*Config, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения конфигурационного файла: %v", err)
	}

	var cfg Config
	// Разбор YAML содержимого в структуру Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка разбора конфигурации: %v", err)
	}

	return &cfg, nil
}
