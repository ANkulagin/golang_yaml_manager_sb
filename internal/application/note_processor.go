package application

import (
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain/entity"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain/service"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/config"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Processor struct {
	Config      *config.Config
	FileRep     domain.FileRepository
	NoteService service.NoteService
	Logger      *logrus.Logger
	sem         chan struct{}
}

func NewProcessor(
	cfg *config.Config,
	fr domain.FileRepository,
	ns service.NoteService,
	logger *logrus.Logger,
) *Processor {
	return &Processor{
		Config:      cfg,
		FileRep:     fr,
		NoteService: ns,
		Logger:      logger,
		sem:         make(chan struct{}, cfg.ConcurrencyLimit),
	}
}

func (p *Processor) Process() error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = p.processDirectory(p.Config.SrcDir, &wg)
	}()
	wg.Wait()
	return nil
}

func (p *Processor) processDirectory(dirPath string, wg *sync.WaitGroup) error {
	p.sem <- struct{}{}
	defer func() { <-p.sem }()

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
			go func(path string) {
				defer wg.Done()
				_ = p.processDirectory(path, wg)
			}(fullPath)
		} else {
			if strings.HasSuffix(file.Name(), ".md") {
				wg.Add(1)
				go func(path string) {
					defer wg.Done()
					p.processFile(path)
				}(fullPath)
			}
		}
	}
	return nil
}

func (p *Processor) processFile(filePath string) {
	p.Logger.Info("file processing" + filePath)
	content, err := p.FileRep.ReadFile(filePath)
	if err != nil {
		p.Logger.Error("file reading error " + filePath + ": " + err.Error())
		return
	}

	note := &entity.Note{
		FilePath:    filePath,
		Content:     content,
		FrontMatter: make(map[string]any),
	}

	if note.HasYaml() {
		if err := note.LoadFrontMatter(); err != nil {
			p.Logger.Error("parsing yaml in file error " + filePath + ": " + err.Error())
			return
		}

		record, err := p.NoteService.ValidateAndUpdate(note)
		if err != nil {
			p.Logger.Error("file validation error " + filePath + ": " + err.Error())
			return
		}
		if record {
			// Генерируем ссылку в формате [[<название файла без расширения>]]
			title := strings.TrimSuffix(filepath.Base(filePath), ".md")
			link := "[[" + title + "]]"
			if err := p.FileRep.AppendToFile(p.Config.ReportFile, link); err != nil {
				p.Logger.Error("Box recording of the report links for file " + filePath + ": " + err.Error())
			} else {
				p.Logger.Info("The link is recorded in the report: " + link)
			}
		}

		if err := p.FileRep.WriteFile(filePath, note.Content); err != nil {
			p.Logger.Error("File recording error " + filePath + ": " + err.Error())
		}
	} else {
		templateContent, err := p.FileRep.ReadFile(p.Config.TemplateDir)
		if err != nil {
			p.Logger.Error("File reading error: " + err.Error())
			return
		}
		note.Content = templateContent + "\n" + note.Content
		//todo При вставке шаблона можно также добавить ссылку в отчёт, если требуется
		if err := p.FileRep.WriteFile(filePath, note.Content); err != nil {
			p.Logger.Error("File recording error " + filePath + ": " + err.Error())
		}
	}
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
