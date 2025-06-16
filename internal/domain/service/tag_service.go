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

	// Регулярное выражение для удаления эмодзи
	emojiRegex := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]`)

	tags := make([]string, 0)
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}

		// Удаляем эмодзи
		cleanPart := emojiRegex.ReplaceAllString(part, "")
		cleanPart = strings.TrimSpace(cleanPart)

		if cleanPart != "" {
			// Преобразуем в snake_case
			tag := toSnakeCase(cleanPart)
			tags = append(tags, tag)
		}
	}

	return tags
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
