package logger

import (
	"fmt"
	"strings"
	_ "time"

	"github.com/sirupsen/logrus"
)

// CustomFormatter определяет форматтер для вывода логов в желаемом формате.
type CustomFormatter struct{}

// Format форматирует запись лога в нужный формат.
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Форматируем время в нужном формате
	timestamp := entry.Time.Format("2006-01-02 15:04:05")

	var levelColor string
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = "\x1b[37m" // белый
	case logrus.InfoLevel:
		levelColor = "\x1b[32m" // зеленый
	case logrus.WarnLevel:
		levelColor = "\x1b[33m" // желтый
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = "\x1b[31m" // красный
	default:
		levelColor = "\x1b[0m" // по умолчанию
	}

	// Формируем строку вывода лога
	var fieldsString string
	for k, v := range entry.Data {
		fieldsString += fmt.Sprintf("%s=%v ", k, v)
	}

	// Сборка окрашенной строки лога
	logMessage := fmt.Sprintf("%s[%s]:%s - %s %s\n",
		levelColor,
		strings.ToUpper(entry.Level.String()),
		timestamp,
		entry.Message,
		fieldsString)
	return []byte(logMessage), nil
}
