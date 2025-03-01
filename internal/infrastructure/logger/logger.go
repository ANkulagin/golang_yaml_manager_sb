package logger

import (
	"github.com/sirupsen/logrus"
	"log"
)

func InitLogger(levelLog string) *logrus.Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	level, err := logrus.ParseLevel(levelLog)
	if err != nil {
		log.Fatalf("Не удалось установить уровень логирования: %v", err)
	}
	logger.SetLevel(level)

	return logger
}
