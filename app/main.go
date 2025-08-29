package main

import (
	"app/bot"
	"app/bot/config"
	"app/bot/storage"
	"log"
)

func main() {
	defer storage.CloseConnection()
	config.LoadAll()

	if err := bot.Start(); err != nil {
		log.Fatalf("Ошибка при запуске бота: %v", err)
	}
}
