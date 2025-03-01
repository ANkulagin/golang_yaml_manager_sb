package domain

//go:generate mockery --case=underscore --dir=. --name=FileRepository --output=../../mocks/repository

type FileRepository interface {
	ReadFile(path string) (string, error)
	WriteFile(path, content string) error
	AppendToFile(path, content string) error
}
