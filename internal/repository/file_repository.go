package repository

import "os"

//go:generate mockery --case=underscore --dir=. --name=FileRepository --output=../../mocks/repository

type FileRepository interface {
	ReadFile(path string) (string, error)
	WriteFile(path, content string) error
	AppendToFile(path, content string) error
}

type fileRepository struct{}

func NewFileRepository() FileRepository {
	return &fileRepository{}
}

func (fr *fileRepository) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (fr *fileRepository) WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), os.ModePerm)
}
func (fr *fileRepository) AppendToFile(path, content string) error {
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
