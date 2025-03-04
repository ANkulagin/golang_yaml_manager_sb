package application

import (
	"bytes"
	"errors"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/config"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/logger"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"testing"

	mocksRepo "github.com/ANkulagin/golang_yaml_manager_sb/mocks/repository"
	mocksService "github.com/ANkulagin/golang_yaml_manager_sb/mocks/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// todo добавлены комменты для пересмотра логики репортов в будущем
func TestProcessor_Process_Success(t *testing.T) {
	tDir, err := os.MkdirTemp("", "proc_test")
	require.NoError(t, err)
	defer os.RemoveAll(tDir)

	subDir := filepath.Join(tDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0755))

	file1 := filepath.Join(tDir, "file1.md")   // есть YAML
	file2 := filepath.Join(tDir, "file2.txt")  // пропустим
	file3 := filepath.Join(subDir, "file3.md") // есть YAML
	file4 := filepath.Join(subDir, "file4.md") // нет YAML

	require.NoError(t, os.WriteFile(file1, []byte("dummy"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("dummy"), 0644))
	require.NoError(t, os.WriteFile(file3, []byte("dummy"), 0644))
	require.NoError(t, os.WriteFile(file4, []byte("dummy"), 0644))

	fileRepoMock := mocksRepo.NewNoteRepository(t)
	noteServiceMock := mocksService.NewNoteService(t)
	loggerInit := logger.InitLogger("info")

	cfg := &config.Config{
		ReportFile:       filepath.Join(tDir, "report.md"),
		TemplateDir:      "template.md",
		ConcurrencyLimit: 2,
		SkipPatterns:     []string{},
		SrcDir:           tDir,
	}

	sut := NewNoteProcessor(cfg.SrcDir, cfg.TemplateDir, cfg.ReportFile, cfg.SkipPatterns, cfg.ConcurrencyLimit, loggerInit, fileRepoMock, noteServiceMock)

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

	content4 := `content4`

	// Моки чтения файлов
	fileRepoMock.On("GetFileContent", file1).Return(content1, nil)
	fileRepoMock.On("GetFileContent", file3).Return(content3, nil)
	fileRepoMock.On("GetFileContent", file4).Return(content4, nil)

	// Когда увидим, что file4 не имеет YAML, код попросит прочитать template.md
	fileRepoMock.On("GetFileContent", "template.md").Return("TEMPLATE_CONTENT", nil)

	// Для файлов с YAML (file1, file3) — вызывается ValidateAndUpsert => true => AddLineToFile => UpdateFileContent
	noteServiceMock.On("ValidateAndUpsert", mock.AnythingOfType("*entity.Note")).Return(true, nil).Twice()

	// => AddLineToFile для file1 и file3
	fileRepoMock.On("AddLineToFile", cfg.ReportFile, "[[file1]]").Return(nil)
	fileRepoMock.On("AddLineToFile", cfg.ReportFile, "[[file3]]").Return(nil)

	// => UpdateFileContent для file1 и file3
	fileRepoMock.On("UpdateFileContent", file1, mock.Anything).Return(nil)
	fileRepoMock.On("UpdateFileContent", file3, mock.Anything).Return(nil)

	// Для file4 без YAML => только вставляем TEMPLATE_CONTENT + "\n" + content4 => и UpdateFileContent
	// Не вызываем ValidateAndUpsert или AddLineToFile
	fileRepoMock.On("UpdateFileContent", file4, mock.Anything).Return(nil)

	// Запуск
	err = sut.Execute()
	require.NoError(t, err)

	fileRepoMock.AssertExpectations(t)
	noteServiceMock.AssertExpectations(t)
}

func TestNoteProcessor_SkipDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skip_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	skipDir := filepath.Join(tmpDir, ".skipMe")
	require.NoError(t, os.Mkdir(skipDir, 0755))

	skipFile := filepath.Join(skipDir, "inside.md")
	require.NoError(t, os.WriteFile(skipFile, []byte("dummy"), 0644))

	fileRepoMock := mocksRepo.NewNoteRepository(t)
	noteServiceMock := mocksService.NewNoteService(t)

	skipPatterns := []string{"."}

	sut := newTestProcessor(t, tmpDir, "template.md", "report.md", skipPatterns, 2, fileRepoMock, noteServiceMock)

	err = sut.Execute()
	require.NoError(t, err)

	// Проверяем, что мы не вызвали моки для skipFile:
	fileRepoMock.AssertNotCalled(t, "GetFileContent", skipFile)
}

func TestNoteProcessor_FileWithoutYaml_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "no_yaml")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("some content"), 0644))

	fileRepoMock := mocksRepo.NewNoteRepository(t)
	noteServiceMock := mocksService.NewNoteService(t)

	tplPath := filepath.Join(tmpDir, "template.md")

	fileRepoMock.On("GetFileContent", filePath).Return("some content", nil)

	fileRepoMock.On("GetFileContent", tplPath).Return("TEMPLATE", nil)

	fileRepoMock.On("UpdateFileContent", filePath, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			updatedContent := args.Get(1).(string)
			require.Equal(t, "TEMPLATE\nsome content", updatedContent)
		})

	noteServiceMock.AssertNotCalled(t, "ValidateAndUpsert", mock.Anything)

	sut := newTestProcessor(t, tmpDir, tplPath, "", nil, 1, fileRepoMock, noteServiceMock)

	err = sut.Execute()
	require.NoError(t, err)
}

func TestNoteProcessor_ValidateUpsert_Closed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "val_closed")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("dummy"), 0644))

	fileRepoMock := mocksRepo.NewNoteRepository(t)
	noteServiceMock := mocksService.NewNoteService(t)

	contentWithYaml := `---
title: test
closed: true
---
Hello
`
	fileRepoMock.On("GetFileContent", filePath).Return(contentWithYaml, nil)
	noteServiceMock.On("ValidateAndUpsert", mock.Anything).Return(false, nil)

	fileRepoMock.AssertNotCalled(t, "AddLineToFile", mock.Anything, mock.Anything)

	fileRepoMock.On("UpdateFileContent", filePath, mock.Anything).Return(nil)

	sut := newTestProcessor(t, tmpDir, "", "", nil, 1, fileRepoMock, noteServiceMock)
	err = sut.Execute()
	require.NoError(t, err)

	fileRepoMock.AssertNotCalled(t, "AddLineToFile", mock.Anything, mock.Anything)
}

func TestNoteProcessor_FileWithoutYaml_ErrorReadingTemplate_Log(t *testing.T) {
	// Если при чтении шаблона ошибка — логируем её, но не падаем

	tmpDir, err := os.MkdirTemp("", "no_yaml_err_tpl")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("some content"), 0644))

	fileRepoMock := mocksRepo.NewNoteRepository(t)
	noteServiceMock := mocksService.NewNoteService(t)

	fileRepoMock.On("GetFileContent", filePath).Return("some content", nil)
	fileRepoMock.On("GetFileContent", "bad_template.md").
		Return("", errors.New("template file not found"))

	sut, logBuf := newTestProcessorWithLoggerBuf(
		t, tmpDir, "bad_template.md", "",
		nil, 1, fileRepoMock, noteServiceMock,
	)

	err = sut.Execute()
	require.NoError(t, err)

	logs := logBuf.String()
	require.Contains(t, logs, "template file not found")
	require.Contains(t, logs, "error handling file")
}

func TestNoteProcessor_FillFrontMatterError_Log(t *testing.T) {
	// Если YAML битый, FillFrontMatter упадёт — мы логируем и идём дальше

	tmpDir, err := os.MkdirTemp("", "front_err")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("dummy"), 0644))

	fileRepoMock := mocksRepo.NewNoteRepository(t)
	noteServiceMock := mocksService.NewNoteService(t)

	brokenYaml := `---
title: 
   - invalid
   ...`
	fileRepoMock.On("GetFileContent", filePath).Return(brokenYaml, nil)

	sut, logBuf := newTestProcessorWithLoggerBuf(
		t, tmpDir, "", "", nil, 1,
		fileRepoMock, noteServiceMock,
	)

	err = sut.Execute()
	require.NoError(t, err)

	logs := logBuf.String()
	require.Contains(t, logs, "incorrectly format front matter") // или любое ваше сообщение
	require.Contains(t, logs, "error handling file")
}

func TestNoteProcessor_ValidateUpsert_Error_Log(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "val_err")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("dummy"), 0644))

	fileRepoMock := mocksRepo.NewNoteRepository(t)
	noteServiceMock := mocksService.NewNoteService(t)

	contentWithYaml := `---
title: test
closed: false
---
Hello
`
	fileRepoMock.On("GetFileContent", filePath).Return(contentWithYaml, nil)
	noteServiceMock.On("ValidateAndUpsert", mock.Anything).
		Return(false, errors.New("some validation error"))

	sut, logBuf := newTestProcessorWithLoggerBuf(
		t, tmpDir, "", "", nil, 1,
		fileRepoMock, noteServiceMock,
	)

	err = sut.Execute()
	require.NoError(t, err)

	logs := logBuf.String()
	require.Contains(t, logs, "some validation error")
	require.Contains(t, logs, "error handling file")
}

func TestNoteProcessor_AddLine_Error_Log(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "add_line_err")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("dummy"), 0644))

	fileRepoMock := mocksRepo.NewNoteRepository(t)
	noteServiceMock := mocksService.NewNoteService(t)

	contentWithYaml := `---
title: test
closed: false
---
something
`
	fileRepoMock.On("GetFileContent", filePath).Return(contentWithYaml, nil)
	noteServiceMock.On("ValidateAndUpsert", mock.Anything).Return(true, nil)

	fileRepoMock.On("AddLineToFile", "report.md", "[[test]]").
		Return(errors.New("add line error"))

	sut, logBuf := newTestProcessorWithLoggerBuf(
		t, tmpDir, "", "report.md", nil, 1,
		fileRepoMock, noteServiceMock,
	)
	err = sut.Execute()
	require.NoError(t, err)

	logs := logBuf.String()
	require.Contains(t, logs, "add line error")
	require.Contains(t, logs, "error handling file")
}

func TestNoteProcessor_UpdateFile_Error_Log(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "upd_err")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("dummy"), 0644))

	fileRepoMock := mocksRepo.NewNoteRepository(t)
	noteServiceMock := mocksService.NewNoteService(t)

	contentWithYaml := `---
title: test
closed: false
---
something
`
	fileRepoMock.On("GetFileContent", filePath).Return(contentWithYaml, nil)
	noteServiceMock.On("ValidateAndUpsert", mock.Anything).Return(true, nil)

	fileRepoMock.On("AddLineToFile", mock.Anything, mock.Anything).Return(nil)

	fileRepoMock.On("UpdateFileContent", filePath, mock.Anything).
		Return(errors.New("update error"))

	sut, logBuf := newTestProcessorWithLoggerBuf(
		t, tmpDir, "", "report.md", nil, 1,
		fileRepoMock, noteServiceMock,
	)
	err = sut.Execute()
	require.NoError(t, err)

	logs := logBuf.String()
	require.Contains(t, logs, "update error")
	require.Contains(t, logs, "error handling file")
}

func newTestProcessor(
	t *testing.T,
	srcDir, tpl, report string,
	skip []string,
	conc int,
	fileRepo *mocksRepo.NoteRepository,
	noteSrv *mocksService.NoteService,
) *NoteProcessor {
	l := logger.InitLogger("debug") // уровень debug, чтобы видеть логи
	return NewNoteProcessor(srcDir, tpl, report, skip, conc, l, fileRepo, noteSrv)
}

func newTestProcessorWithLoggerBuf(
	t *testing.T,
	srcDir, tplPath, reportPath string,
	skipPatterns []string,
	concurrency int,
	fileRepo *mocksRepo.NoteRepository,
	noteSrv *mocksService.NoteService,
) (*NoteProcessor, *bytes.Buffer) {

	var logBuf bytes.Buffer
	testLogger := logrus.New()
	testLogger.SetOutput(&logBuf)
	testLogger.SetLevel(logrus.DebugLevel)

	processor := NewNoteProcessor(
		srcDir, tplPath, reportPath,
		skipPatterns, concurrency,
		testLogger, fileRepo, noteSrv,
	)
	return processor, &logBuf
}
