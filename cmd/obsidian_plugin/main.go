package main

import (
	"log"

	"github.com/ANkulagin/golang_yaml_manager_sb/internal/application"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain/service"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/config"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/logger"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/repository"
)

func main() {
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Error load config: %v", err)
	}

	logg := logger.InitLogger(cfg.LogLevel)
	logg.Infof("Level log %s", cfg.LogLevel)

	noteRepo := repository.NewFileRepository()
	noteSrv := service.NewNoteService()
	tagSrv := service.NewTagService()

	processor := application.NewNoteProcessor(
		cfg.SrcDir,
		cfg.TemplateDir,
		cfg.ReportFile,
		cfg.SkipPatterns,
		cfg.ConcurrencyLimit,
		logg,
		noteRepo,
		noteSrv,
		tagSrv,
	)

	if err := processor.Execute(); err != nil {
		logg.Errorf("Error executing note processing: %v", err)
	}
}
