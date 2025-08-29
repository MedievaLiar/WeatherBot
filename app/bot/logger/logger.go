package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"app/bot/models"
)

var (
	logFile *os.File
	logger  *log.Logger
)

// Инициализирует систему логирования
func Init() error {
	logsDir := "bot/logger/logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию логов: %w", err)
	}

	// Создаем файл лога с текущей датой
	logPath := filepath.Join(logsDir, fmt.Sprintf("bot_%s.log", time.Now().Format("2006-01-02")))

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл лога: %w", err)
	}

	logFile = file
	logger = log.New(file, "", log.LstdFlags|log.Lshortfile)

	logger.Printf("Система логирования инициализирована")
	return nil
}

// Логирует добавление пользователя
func LogUserAdded(user models.UserData) {
	message := fmt.Sprintf("ДОБАВЛЕН ПОЛЬЗОВАТЕЛЬ: ID=%d, Username=@%s, CurrentCity=%s, ForecastCity=%s, WantDaily=%t",
		user.ID, user.Username, user.CurrentCity, user.ForecastCity, user.WantDaily)

	logger.Println(message)
	fmt.Println(message) // Также выводим в консоль
}

// Логирует обновление пользователя
func LogUserUpdated(user models.UserData) {
	message := fmt.Sprintf("ОБНОВЛЕН ПОЛЬЗОВАТЕЛЬ: ID=%d, Username=@%s, CurrentCity=%s, ForecastCity=%s, WantDaily=%t",
		user.ID, user.Username, user.CurrentCity, user.ForecastCity, user.WantDaily)

	logger.Println(message)
	fmt.Println(message)
}

// Close закрывает файл лога
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}
