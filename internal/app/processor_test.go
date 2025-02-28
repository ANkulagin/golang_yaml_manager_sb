package app

import (
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/config"
	mocksRepo "github.com/ANkulagin/golang_yaml_manager_sb/mocks/repository"
	mocksService "github.com/ANkulagin/golang_yaml_manager_sb/mocks/service"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestProcessor_ProcessDirectory_Success(t *testing.T) {
	tDir, err := os.MkdirTemp("", "proc_test")
	require.NoError(t, err)
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tDir)

	subDir := filepath.Join(tDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0755))

	file1 := filepath.Join(tDir, "file1.md")
	file2 := filepath.Join(tDir, "file2.txt")
	file3 := filepath.Join(subDir, "file3.md")

	// Физически файлы могут быть пустыми, т.к. чтение происходит через моки.
	require.NoError(t, os.WriteFile(file1, []byte("dummy"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("dummy"), 0644))
	require.NoError(t, os.WriteFile(file3, []byte("dummy"), 0644))

	fileRepoMock := mocksRepo.NewFileRepository(t)
	noteServiceMock := mocksService.NewNoteService(t)

	cfg := &config.Config{
		ReportFile:       filepath.Join(tDir, "report.md"),
		TemplateDir:      "template.md",
		ConcurrencyLimit: 2,
		SkipPatterns:     []string{},
	}

	sut := NewProcessor(cfg, fileRepoMock, noteServiceMock)

	//	content1 := `---
	//title: file1
	//closed: false
	//---
	//content1`
	//	content3 := `---
	//title: file3
	//closed: false
	//---
	//content3`
	//
	//	fileRepoMock.On("ReadFile", file1).Return(content1, nil)
	//	fileRepoMock.On("ReadFile", file3).Return(content3, nil)

	// Используем WaitGroup для ожидания завершения всех горутин.
	var wg sync.WaitGroup

	err = sut.ProcessDirectory(tDir, &wg)
	require.NoError(t, err)

	// Ожидаем завершения всех горутин.
	wg.Wait()

}
