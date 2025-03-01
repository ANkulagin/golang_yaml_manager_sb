package application

import (
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/config"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/logger"
	"os"
	"path/filepath"
	"testing"

	mocksRepo "github.com/ANkulagin/golang_yaml_manager_sb/mocks/repository"
	mocksService "github.com/ANkulagin/golang_yaml_manager_sb/mocks/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProcessor_Process_Success(t *testing.T) {
	tDir, err := os.MkdirTemp("", "proc_test")
	require.NoError(t, err)
	defer os.RemoveAll(tDir)

	subDir := filepath.Join(tDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0755))

	file1 := filepath.Join(tDir, "file1.md")
	file2 := filepath.Join(tDir, "file2.txt")
	file3 := filepath.Join(subDir, "file3.md")

	require.NoError(t, os.WriteFile(file1, []byte("dummy"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("dummy"), 0644))
	require.NoError(t, os.WriteFile(file3, []byte("dummy"), 0644))

	fileRepoMock := mocksRepo.NewFileRepository(t)
	noteServiceMock := mocksService.NewNoteService(t)
	loggerInit := logger.InitLogger("info")

	cfg := &config.Config{
		ReportFile:       filepath.Join(tDir, "report.md"),
		TemplateDir:      "template.md",
		ConcurrencyLimit: 2,
		SkipPatterns:     []string{},
		SrcDir:           tDir,
	}

	sut := NewProcessor(cfg, fileRepoMock, noteServiceMock, loggerInit)

	content1 := `---
title: file1
closed: false
---
content1`
	content3 := `---
title: file3
closed: false
---
content3`

	fileRepoMock.On("ReadFile", file1).Return(content1, nil)
	fileRepoMock.On("ReadFile", file3).Return(content3, nil)

	noteServiceMock.
		On("ValidateAndUpdate", mock.AnythingOfType("*entity.Note")).
		Return(true, nil).Twice()

	fileRepoMock.
		On("AppendToFile", cfg.ReportFile, "[[file1]]").
		Return(nil)
	fileRepoMock.
		On("AppendToFile", cfg.ReportFile, "[[file3]]").
		Return(nil)

	fileRepoMock.
		On("WriteFile", file1, mock.Anything).
		Return(nil)
	fileRepoMock.
		On("WriteFile", file3, mock.Anything).
		Return(nil)

	err = sut.Process()
	require.NoError(t, err)

	fileRepoMock.AssertExpectations(t)
	noteServiceMock.AssertExpectations(t)
}
