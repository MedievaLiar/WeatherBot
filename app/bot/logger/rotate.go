package logger

import (
	"os"
	"path/filepath"
	"time"
)

// StartRotation запускает ротацию логов
func StartRotation() {
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // Проверяем каждые 24 часа
		defer ticker.Stop()

		for range ticker.C {
			cleanOldLogs()
		}
	}()
}

// cleanOldLogs удаляет логи старше 5 дней
func cleanOldLogs() {
	logsDir := "./storage/logs"
	files, err := os.ReadDir(logsDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(logsDir, file.Name())
		info, err := file.Info()
		if err != nil {
			continue
		}

		// Удаляем файлы старше 5 дней
		if time.Since(info.ModTime()) > 5*24*time.Hour {
			os.Remove(filePath)
		}
	}
}
