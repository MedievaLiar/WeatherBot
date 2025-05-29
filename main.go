package main

import (
	"log"
	"weather_bot/bot"
	"weather_bot/config"
)

func main() {
	config.LoadAll()

	if err := bot.Start(); err != nil {
		log.Fatalf("Ошибка при запуске бота: %v", err)
	}
}
