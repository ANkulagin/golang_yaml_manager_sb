package service

import (
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain/entity"
)

//go:generate mockery --case=underscore --dir=. --name=NoteService --output=../../../mocks/service

type NoteService interface {
	ValidateAndUpsert(note *entity.Note) (bool, error)
}

type noteService struct{}

func NewsNoteService() NoteService {
	return &noteService{}
}

// ValidateAndUpsert проверяет наличие поля 'closed' и обновляет YAML-шапку при необходимости.
// Возвращает true, если заметку следует добавить в отчёт.
func (ns *noteService) ValidateAndUpsert(note *entity.Note) (bool, error) {
	closedVal, exists := note.FrontMatter["closed"]
	if exists {
		if closed, ok := closedVal.(bool); ok && closed {
			// Заметка помечена закрытой – пропускаем обработку.
			return false, nil
		}
		// Если значение false, отмечаем для отчёта.
		return true, nil
	}

	// Если поле 'closed' отсутствует, добавляем его со значением false.
	note.FrontMatter["closed"] = false
	if err := note.UpdateFrontMatter(); err != nil {
		return false, err
	}
	// Отмечаем заметку для добавления в отчёт.
	return true, nil
}
