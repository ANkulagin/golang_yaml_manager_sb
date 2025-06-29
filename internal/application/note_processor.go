package application

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain/entity"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain/service"
)

type NoteProcessor struct {
	SrcDir           string
	TemplatePath     string
	ReportPath       string
	SkipPatterns     []string
	ConcurrencyLimit int

	Logger      *logrus.Logger
	NoteRepo    domain.NoteRepository
	NoteService service.NoteService
	TagService  service.TagService

	sem         chan struct{}
	reportMutex sync.Mutex
	reportLinks []string
}

func NewNoteProcessor(
	srcDir, templatePath, reportPath string,
	skipPatterns []string,
	concurrencyLimit int,
	logger *logrus.Logger,
	noteRepo domain.NoteRepository,
	noteService service.NoteService,
	tagService service.TagService,
) *NoteProcessor {
	return &NoteProcessor{
		SrcDir:           srcDir,
		TemplatePath:     templatePath,
		ReportPath:       reportPath,
		SkipPatterns:     skipPatterns,
		ConcurrencyLimit: concurrencyLimit,
		Logger:           logger,
		NoteRepo:         noteRepo,
		NoteService:      noteService,
		TagService:       tagService,
		sem:              make(chan struct{}, concurrencyLimit),
		reportLinks:      make([]string, 0),
	}
}

func (p *NoteProcessor) Execute() error {
	// Очищаем файл отчёта перед началом
	if err := p.NoteRepo.ClearFile(p.ReportPath); err != nil {
		p.Logger.Warnf("failed to clear report file: %v", err)
	}

	var wg sync.WaitGroup
	_ = p.handleDirectory(p.SrcDir, &wg)
	wg.Wait()

	// Записываем все ссылки в отчёт
	for _, link := range p.reportLinks {
		if err := p.NoteRepo.AddLineToFile(p.ReportPath, link); err != nil {
			p.Logger.Errorf("failed to add link to report: %v", err)
		}
	}

	return nil
}

func (p *NoteProcessor) handleDirectory(dirPath string, wg *sync.WaitGroup) error {
	p.sem <- struct{}{}
	defer func() { <-p.sem }()

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() && p.checkSkipDirectory(f.Name()) {
			continue
		}

		fullPath := filepath.Join(dirPath, f.Name())
		if f.IsDir() {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()
				_ = p.handleDirectory(path, wg)
			}(fullPath)
		} else {
			if strings.HasSuffix(f.Name(), ".md") {
				wg.Add(1)
				go func(path string) {
					defer wg.Done()
					if err := p.handleFile(path); err != nil {
						p.Logger.Errorf("error handling file %s: %v", path, err)
					}
				}(fullPath)
			}
		}
	}
	return nil
}

func (p *NoteProcessor) handleFile(filePath string) error {
	p.Logger.Infof("handling file: %s", filePath)

	content, err := p.NoteRepo.GetFileContent(filePath)
	if err != nil {
		return err
	}

	note := &entity.Note{
		FilePath:    filePath,
		Content:     content,
		FrontMatter: make(map[string]any),
	}

	if !note.CheckHasYaml() {
		tpl, err := p.NoteRepo.GetFileContent(p.TemplatePath)
		if err != nil {
			return err
		}
		note.Content = tpl + "\n" + note.Content
		return p.NoteRepo.UpdateFileContent(filePath, note.Content)
	}

	if err := note.FillFrontMatter(); err != nil {
		return err
	}

	// Извлекаем теги из пути и обновляем
	pathTags := p.TagService.ExtractTagsFromPath(filePath, p.SrcDir)
	if len(pathTags) > 0 {
		if err := p.NoteService.UpdateTags(note, pathTags); err != nil {
			p.Logger.Errorf("failed to update tags for %s: %v", filePath, err)
		}
	}

	shouldReport, err := p.NoteService.ValidateAndUpsert(note)
	if err != nil {
		return err
	}

	if shouldReport {
		title := strings.TrimSuffix(filepath.Base(filePath), ".md")
		link := "[[" + title + "]]"

		p.reportMutex.Lock()
		p.reportLinks = append(p.reportLinks, link)
		p.reportMutex.Unlock()

		p.Logger.Infof("will add link to report: %s", link)
	}

	return p.NoteRepo.UpdateFileContent(filePath, note.Content)
}

func (p *NoteProcessor) checkSkipDirectory(dirname string) bool {
	for _, prefix := range p.SkipPatterns {
		if strings.HasPrefix(dirname, prefix) {
			return true
		}
	}
	return false
}
