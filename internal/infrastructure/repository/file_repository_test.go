package repository

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestFileRepository_ReadFile(t *testing.T) {
	fileRepo := NewFileRepository()
	tempFile, err := os.CreateTemp("", "testfile*.txt")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	content := "Hello, Go!"
	err = os.WriteFile(tempFile.Name(), []byte(content), os.ModePerm)
	require.NoError(t, err)

	result, err := fileRepo.GetFileContent(tempFile.Name())
	require.NoError(t, err)
	require.Equal(t, content, result)
}

func TestFileRepository_WriteFile(t *testing.T) {
	fileRepo := NewFileRepository()
	tempFile, err := os.CreateTemp("", "testfile*.txt")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	content := "Write this to file"
	err = fileRepo.UpdateFileContent(tempFile.Name(), content)
	require.NoError(t, err)

	data, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)
	require.Equal(t, content, string(data))
}

func TestFileRepository_AppendToFile(t *testing.T) {
	fileRepo := NewFileRepository()
	tempFile, err := os.CreateTemp("", "testfile*.txt")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	initialContent := "Initial content\n"
	appendContent := "Appended content"
	err = os.WriteFile(tempFile.Name(), []byte(initialContent), os.ModePerm)
	require.NoError(t, err)
	err = fileRepo.AddLineToFile(tempFile.Name(), appendContent)
	require.NoError(t, err)

	data, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)
	require.Equal(t, initialContent+appendContent+"\n", string(data))
}
