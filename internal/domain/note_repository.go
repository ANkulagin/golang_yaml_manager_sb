package domain

//go:generate mockery --case=underscore --dir=. --name=NoteRepository --output=../../mocks/repository

type NoteRepository interface {
	GetFileContent(path string) (string, error)
	UpdateFileContent(path, content string) error
	AddLineToFile(path, content string) error
}
