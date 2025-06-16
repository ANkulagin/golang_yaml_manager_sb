package service

import (
	"path/filepath"
	"regexp"
	"strings"
)

type TagService interface {
	ExtractTagsFromPath(filePath string, srcDir string) []string
}

type tagService struct{}

func NewTagService() TagService {
	return &tagService{}
}

func (ts *tagService) ExtractTagsFromPath(filePath string, srcDir string) []string {
	// Получаем относительный путь
	relPath, err := filepath.Rel(srcDir, filePath)
	if err != nil {
		return []string{}
	}

	// Получаем директорию без имени файла
	dir := filepath.Dir(relPath)

	// Разбиваем путь на части
	parts := strings.Split(dir, string(filepath.Separator))

	tags := make([]string, 0)
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}

		// Удаляем эмодзи и очищаем строку
		cleanPart := removeEmojis(part)
		cleanPart = strings.TrimSpace(cleanPart)

		if cleanPart != "" {
			// Преобразуем в snake_case
			tag := toSnakeCase(cleanPart)
			tags = append(tags, tag)
		}
	}

	return tags
}

// removeEmojis удаляет все эмодзи из строки
func removeEmojis(s string) string {
	// Создаём новую строку без эмодзи
	var result strings.Builder

	for _, r := range s {
		// Проверяем, является ли символ эмодзи
		if !isEmoji(r) {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// isEmoji проверяет, является ли руна эмодзи
func isEmoji(r rune) bool {
	// Основные блоки эмодзи в Unicode
	return (r >= 0x1F600 && r <= 0x1F64F) || // Эмоции
		(r >= 0x1F300 && r <= 0x1F5FF) || // Символы и пиктограммы
		(r >= 0x1F680 && r <= 0x1F6FF) || // Транспорт и карты
		(r >= 0x1F1E0 && r <= 0x1F1FF) || // Флаги
		(r >= 0x2600 && r <= 0x26FF) || // Разные символы
		(r >= 0x2700 && r <= 0x27BF) || // Дингбаты
		(r >= 0xFE00 && r <= 0xFE0F) || // Селекторы вариантов
		(r >= 0x1F900 && r <= 0x1F9FF) || // Дополнительные символы
		(r >= 0x1FA70 && r <= 0x1FAFF) || // Символы и пиктограммы расширенные
		(r >= 0xE0000 && r <= 0xE007F) // Теги
}

func toSnakeCase(s string) string {
	// Заменяем пробелы на подчёркивания
	s = strings.ReplaceAll(s, " ", "_")

	// Приводим к нижнему регистру
	s = strings.ToLower(s)

	// Удаляем множественные подчёркивания
	re := regexp.MustCompile(`_+`)
	s = re.ReplaceAllString(s, "_")

	// Удаляем подчёркивания в начале и конце
	s = strings.Trim(s, "_")

	return s
}
