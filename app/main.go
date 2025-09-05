package main

import (
	"log"

	"app/bot"
	"app/bot/config"
	"app/bot/logger"
)

func main() {
	if err := logger.Init(); err != nil {
		log.Fatalf("Не удалось инициализировать логгер: %v", err)
	}
	defer logger.Close()

	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC: %v", r)
			log.Fatalf("PANIC: %v", r)
		}
	}()

	config.LoadAll()

	if err := bot.Start(); err != nil {
		logger.Error("Критическая ошибка: %v", err)
		log.Fatalf("Ошибка при запуске бота: %v", err)
	}
}
