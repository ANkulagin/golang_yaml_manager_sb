package main

import (
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/app"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/config"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/repository"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/service"
	"github.com/ANkulagin/golang_yaml_manager_sb/pkg/logger"
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

	p := app.NewProcessor(cfg, fr, ns, initLogger)

	if err := p.Process(); err != nil {
		initLogger.Error("error: ", err)
	}
}
