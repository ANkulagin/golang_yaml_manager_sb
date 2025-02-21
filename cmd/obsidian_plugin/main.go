package main

import (
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/config"
	"github.com/ANkulagin/golang_yaml_manager_sb/pkg/logger"
	"log"
)

var configPath = "configs/config.yaml"

func main() {
	logger.InitLogger()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error load config: %v", err)
	}

	logger.SetLevel(cfg.LogLevel)
}
