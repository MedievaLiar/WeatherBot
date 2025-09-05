package logger

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// StartRotation запускает ротацию логов
func StartRotation() {
	go func() {
		cleanOldLogs()

		ticker := time.NewTicker(24 * time.Hour) // Проверяем каждые 24 часа
		defer ticker.Stop()

		for range ticker.C {
			cleanOldLogs()
		}
	}()
}

// удаляет логи старше 5 дней
func cleanOldLogs() {
	logsDir := "bot/logger/logs"
	files, err := os.ReadDir(logsDir)
	if err != nil {
		log.Printf("Ошибка чтения директории логов: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(logsDir, file.Name())

		// получаем метаданные файла
		info, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Ошибка получения информации о файле %s: %v", file.Name(), err)
			continue
		}

		// удаляем файлы старше 5 дней
		if time.Since(info.ModTime()) > 5*24*time.Hour { // время последнего изменения
			if err := os.Remove(filePath); err != nil {
				log.Printf("Ошибка удаления файла %s: %v", file.Name(), err)
			}
		}
	}
}
