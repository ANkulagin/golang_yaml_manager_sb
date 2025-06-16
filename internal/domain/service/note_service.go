package service

import (
	"github.com/ANkulagin/golang_yaml_manager_sb/internal/domain/entity"
)

type NoteService interface {
	ValidateAndUpsert(note *entity.Note) (bool, error)
	UpdateTags(note *entity.Note, newTags []string) error
}

type noteService struct{}

func NewNoteService() NoteService {
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

func (ns *noteService) UpdateTags(note *entity.Note, newTags []string) error {
	existingTags := []string{}

	// Получаем существующие теги
	if tags, exists := note.FrontMatter["tags"]; exists {
		switch v := tags.(type) {
		case []interface{}:
			for _, tag := range v {
				if strTag, ok := tag.(string); ok {
					existingTags = append(existingTags, strTag)
				}
			}
		case []string:
			existingTags = v
		}
	}

	// Создаём map для быстрой проверки существования
	tagMap := make(map[string]bool)
	for _, tag := range existingTags {
		tagMap[tag] = true
	}

	// Добавляем новые теги
	for _, tag := range newTags {
		if !tagMap[tag] {
			existingTags = append(existingTags, tag)
		}
	}

	// Обновляем теги в FrontMatter
	note.FrontMatter["tags"] = existingTags

	return note.UpdateFrontMatter()
}
