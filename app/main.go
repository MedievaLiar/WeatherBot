package main

import (
	"app/bot"
	"app/bot/config"
	"log"
)

func main() {
	config.LoadAll()

	if err := bot.Start(); err != nil {
		log.Fatalf("Ошибка при запуске бота: %v", err)
	}
}
