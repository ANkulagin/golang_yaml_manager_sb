package application

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain/service"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/config"
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/infrastructure/logger"
	"github.com/sirupsen/logrus"

	"github.com/ANkulagin/golang_yaml_manager_sb/mocks"
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
	file4 := filepath.Join(subDir, "file4.md")

	require.NoError(t, os.WriteFile(file1, []byte("dummy"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("dummy"), 0644))
	require.NoError(t, os.WriteFile(file3, []byte("dummy"), 0644))
	require.NoError(t, os.WriteFile(file4, []byte("dummy"), 0644))

	fileRepoMock := mocks.NewNoteRepository(t)
	noteServiceMock := mocks.NewNoteService(t)
	tagSrv := service.NewTagService()
	loggerInit := logger.InitLogger("info")

	cfg := &config.Config{
		ReportFile:       filepath.Join(tDir, "report.md"),
		ReportEnabled:    true,
		TemplateDir:      "template.md",
		ConcurrencyLimit: 2,
		SkipPatterns:     []string{},
		SrcDir:           tDir,
	}

	sut := NewNoteProcessor(cfg.SrcDir, cfg.TemplateDir, cfg.ReportFile, cfg.ReportEnabled, cfg.SkipPatterns, cfg.ConcurrencyLimit, loggerInit, fileRepoMock, noteServiceMock, tagSrv)

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

	fileRepoMock.On("ClearFile", cfg.ReportFile).Return(nil)

	fileRepoMock.On("GetFileContent", file1).Return(content1, nil)
	fileRepoMock.On("GetFileContent", file3).Return(content3, nil)
	fileRepoMock.On("GetFileContent", file4).Return(content4, nil)

	fileRepoMock.On("GetFileContent", "template.md").Return("TEMPLATE_CONTENT", nil)

	noteServiceMock.On("ValidateAndUpsert", mock.AnythingOfType("*entity.Note")).Return(true, nil).Twice()
	noteServiceMock.On("UpdateTags", mock.AnythingOfType("*entity.Note"), []string{"subdir"}).Return(nil).Once()

	fileRepoMock.On("AddLineToFile", cfg.ReportFile, "[[file1]]").Return(nil)
	fileRepoMock.On("AddLineToFile", cfg.ReportFile, "[[file3]]").Return(nil)

	fileRepoMock.On("UpdateFileContent", file1, mock.Anything).Return(nil)
	fileRepoMock.On("UpdateFileContent", file3, mock.Anything).Return(nil)

	fileRepoMock.On("UpdateFileContent", file4, mock.Anything).Return(nil)

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

	fileRepoMock := mocks.NewNoteRepository(t)
	noteServiceMock := mocks.NewNoteService(t)

	skipPatterns := []string{"."}

	sut := newTestProcessor(t, tmpDir, "template.md", "report.md", false, skipPatterns, 2, fileRepoMock, noteServiceMock)

	err = sut.Execute()
	require.NoError(t, err)

	fileRepoMock.AssertNotCalled(t, "GetFileContent", skipFile)
}

func TestNoteProcessor_FileWithoutYaml_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "no_yaml")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("some content"), 0644))

	fileRepoMock := mocks.NewNoteRepository(t)
	noteServiceMock := mocks.NewNoteService(t)

	tplPath := filepath.Join(tmpDir, "template.md")

	fileRepoMock.On("GetFileContent", filePath).Return("some content", nil)
	fileRepoMock.On("GetFileContent", tplPath).Return("TEMPLATE", nil)

	fileRepoMock.On("UpdateFileContent", filePath, mock.Anything).
		Return(nil).
		Run(func(args mock.Arguments) {
			updatedContent := args.Get(1).(string)
			require.Equal(t, "TEMPLATE\nsome content", updatedContent)
		})

	sut := newTestProcessor(t, tmpDir, tplPath, "", false, nil, 1, fileRepoMock, noteServiceMock)

	err = sut.Execute()
	require.NoError(t, err)
}

func TestNoteProcessor_ValidateUpsert_Closed(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "val_closed")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("dummy"), 0644))

	fileRepoMock := mocks.NewNoteRepository(t)
	noteServiceMock := mocks.NewNoteService(t)

	contentWithYaml := `---
title: test
closed: true
---
Hello
`
	fileRepoMock.On("GetFileContent", filePath).Return(contentWithYaml, nil)
	noteServiceMock.On("ValidateAndUpsert", mock.Anything).Return(false, nil)
	fileRepoMock.On("UpdateFileContent", filePath, mock.Anything).Return(nil)

	sut := newTestProcessor(t, tmpDir, "", "", false, nil, 1, fileRepoMock, noteServiceMock)
	err = sut.Execute()
	require.NoError(t, err)

	fileRepoMock.AssertNotCalled(t, "AddLineToFile", mock.Anything, mock.Anything)
}

func TestNoteProcessor_FileWithoutYaml_ErrorReadingTemplate_Log(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "no_yaml_err_tpl")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("some content"), 0644))

	fileRepoMock := mocks.NewNoteRepository(t)
	noteServiceMock := mocks.NewNoteService(t)

	fileRepoMock.On("GetFileContent", filePath).Return("some content", nil)
	fileRepoMock.On("GetFileContent", "bad_template.md").
		Return("", errors.New("template file not found"))

	sut, logBuf := newTestProcessorWithLoggerBuf(
		t, tmpDir, "bad_template.md", "", false,
		nil, 1, fileRepoMock, noteServiceMock,
	)

	err = sut.Execute()
	require.NoError(t, err)

	logs := logBuf.String()
	require.Contains(t, logs, "template file not found")
	require.Contains(t, logs, "error handling file")
}

func TestNoteProcessor_FillFrontMatterError_Log(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "front_err")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("dummy"), 0644))

	fileRepoMock := mocks.NewNoteRepository(t)
	noteServiceMock := mocks.NewNoteService(t)

	brokenYaml := `---
title:
   - invalid
   ...`
	fileRepoMock.On("GetFileContent", filePath).Return(brokenYaml, nil)

	sut, logBuf := newTestProcessorWithLoggerBuf(
		t, tmpDir, "", "", false, nil, 1,
		fileRepoMock, noteServiceMock,
	)

	err = sut.Execute()
	require.NoError(t, err)

	logs := logBuf.String()
	require.Contains(t, logs, "incorrectly format front matter")
	require.Contains(t, logs, "error handling file")
}

func TestNoteProcessor_ValidateUpsert_Error_Log(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "val_err")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("dummy"), 0644))

	fileRepoMock := mocks.NewNoteRepository(t)
	noteServiceMock := mocks.NewNoteService(t)

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
		t, tmpDir, "", "", false, nil, 1,
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

	fileRepoMock := mocks.NewNoteRepository(t)
	noteServiceMock := mocks.NewNoteService(t)

	contentWithYaml := `---
title: test
closed: false
---
something
`
	fileRepoMock.On("ClearFile", "report.md").Return(nil)
	fileRepoMock.On("GetFileContent", filePath).Return(contentWithYaml, nil)
	noteServiceMock.On("ValidateAndUpsert", mock.Anything).Return(true, nil)
	fileRepoMock.On("UpdateFileContent", filePath, mock.Anything).Return(nil)
	fileRepoMock.On("AddLineToFile", "report.md", "[[test]]").
		Return(errors.New("add line error"))

	sut, logBuf := newTestProcessorWithLoggerBuf(
		t, tmpDir, "", "report.md", true, nil, 1,
		fileRepoMock, noteServiceMock,
	)
	err = sut.Execute()
	require.NoError(t, err)

	logs := logBuf.String()
	require.Contains(t, logs, "add line error")
	require.Contains(t, logs, "failed to add link to report")
}

func TestNoteProcessor_UpdateFile_Error_Log(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "upd_err")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("dummy"), 0644))

	fileRepoMock := mocks.NewNoteRepository(t)
	noteServiceMock := mocks.NewNoteService(t)

	contentWithYaml := `---
title: test
closed: false
---
something
`
	fileRepoMock.On("ClearFile", "report.md").Return(nil)
	fileRepoMock.On("GetFileContent", filePath).Return(contentWithYaml, nil)
	noteServiceMock.On("ValidateAndUpsert", mock.Anything).Return(true, nil)
	fileRepoMock.On("AddLineToFile", mock.Anything, mock.Anything).Return(nil)
	fileRepoMock.On("UpdateFileContent", filePath, mock.Anything).
		Return(errors.New("update error"))

	sut, logBuf := newTestProcessorWithLoggerBuf(
		t, tmpDir, "", "report.md", true, nil, 1,
		fileRepoMock, noteServiceMock,
	)
	err = sut.Execute()
	require.NoError(t, err)

	logs := logBuf.String()
	require.Contains(t, logs, "update error")
	require.Contains(t, logs, "error handling file")
}

func TestNoteProcessor_ReportDisabled_SkipsReportIO(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "report_disabled")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(filePath, []byte("dummy"), 0644))

	fileRepoMock := mocks.NewNoteRepository(t)
	noteServiceMock := mocks.NewNoteService(t)

	contentWithYaml := `---
title: test
closed: false
---
body
`
	fileRepoMock.On("GetFileContent", filePath).Return(contentWithYaml, nil)
	noteServiceMock.On("ValidateAndUpsert", mock.Anything).Return(true, nil)
	fileRepoMock.On("UpdateFileContent", filePath, mock.Anything).Return(nil)

	sut := newTestProcessor(t, tmpDir, "", "report.md", false, nil, 1, fileRepoMock, noteServiceMock)
	err = sut.Execute()
	require.NoError(t, err)

	fileRepoMock.AssertNotCalled(t, "ClearFile", mock.Anything)
	fileRepoMock.AssertNotCalled(t, "AddLineToFile", mock.Anything, mock.Anything)
}

func newTestProcessor(
	t *testing.T,
	srcDir, tpl, report string,
	reportEnabled bool,
	skip []string,
	conc int,
	fileRepo *mocks.NoteRepository,
	noteSrv *mocks.NoteService,
) *NoteProcessor {
	l := logger.InitLogger("debug")
	return NewNoteProcessor(srcDir, tpl, report, reportEnabled, skip, conc, l, fileRepo, noteSrv, service.NewTagService())
}

func newTestProcessorWithLoggerBuf(
	t *testing.T,
	srcDir, tplPath, reportPath string,
	reportEnabled bool,
	skipPatterns []string,
	concurrency int,
	fileRepo *mocks.NoteRepository,
	noteSrv *mocks.NoteService,
) (*NoteProcessor, *bytes.Buffer) {

	var logBuf bytes.Buffer
	testLogger := logrus.New()
	testLogger.SetOutput(&logBuf)
	testLogger.SetLevel(logrus.DebugLevel)

	processor := NewNoteProcessor(
		srcDir, tplPath, reportPath, reportEnabled,
		skipPatterns, concurrency,
		testLogger, fileRepo, noteSrv, service.NewTagService(),
	)
	return processor, &logBuf
}
