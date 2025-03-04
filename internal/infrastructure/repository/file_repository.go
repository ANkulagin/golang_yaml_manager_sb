package repository

import (
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain"
	"os"
)

type fileRepository struct{}

func NewFileRepository() domain.NoteRepository {
	return &fileRepository{}
}

func (fr *fileRepository) GetFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (fr *fileRepository) UpdateFileContent(path, content string) error {
	return os.WriteFile(path, []byte(content), os.ModePerm)
}
func (fr *fileRepository) AddLineToFile(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	if _, err := f.WriteString(content + "\n"); err != nil {
		return err
	}

	return nil
}
