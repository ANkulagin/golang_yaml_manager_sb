package main

import (
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/application"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain/service"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/config"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/logger"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/repository"
	"log"
)

var configPath = "configs/config.yaml"

func main() {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error load config: %v", err)
	}

	initLogger := logger.InitLogger(cfg.LogLevel)
	initLogger.Infof("Level log %s", cfg.LogLevel)

	fr := repository.NewFileRepository()
	ns := service.NewsNoteService()

	p := application.NewProcessor(cfg, fr, ns, initLogger)

	if err := p.Process(); err != nil {
		initLogger.Error("error: ", err)
	}
}
