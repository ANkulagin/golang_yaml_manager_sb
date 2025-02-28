package app

import (
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/config"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/repository"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/service"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Processor struct {
	Config      *config.Config
	FileRep     repository.FileRepository
	NoteService service.NoteService
	sem         chan struct{}
}

func NewProcessor(
	cfg *config.Config,
	fr repository.FileRepository,
	ns service.NoteService,
) *Processor {
	return &Processor{
		Config:      cfg,
		FileRep:     fr,
		NoteService: ns,
		sem:         make(chan struct{}, cfg.ConcurrencyLimit),
	}
}

func (p *Processor) ProcessDirectory(dirPath string, wg *sync.WaitGroup) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() && p.shouldSkipDirectory(file.Name()) {
			continue
		}

		fullPath := filepath.Join(dirPath, file.Name())

		if file.IsDir() {
			wg.Add(1)
			p.sem <- struct{}{}

			go func(path string) {
				defer wg.Done()
				defer func() { <-p.sem }()

				_ = p.ProcessDirectory(path, wg) //todo
			}(fullPath)
		} else {
			if strings.HasSuffix(file.Name(), ".md") {
				wg.Add(1)
				p.sem <- struct{}{}

				go func(path string) {
					defer wg.Done()
					defer func() { <-p.sem }()

					p.processFile(path)
				}(fullPath)
			}
		}
	}
	return nil
}

func (p *Processor) processFile(filePath string) {

}

// shouldSkipDirectory возвращает true, если имя директории начинается с одного из указанных префиксов.
func (p *Processor) shouldSkipDirectory(directoryName string) bool {
	for _, prefix := range p.Config.SkipPatterns {
		if strings.HasPrefix(directoryName, prefix) {
			return true
		}
	}
	return false
}
