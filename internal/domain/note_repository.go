package domain

type NoteRepository interface {
	GetFileContent(path string) (string, error)
	UpdateFileContent(path, content string) error
	AddLineToFile(path, content string) error
	ClearFile(path string) error
}
