package main

import (
	"app/bot/config"
	"app/bot/storage/aggregator"
	"fmt"
	"log"
)

func main() {
	city := "Москва"
	config.LoadAll()

	weatherStr, err := aggregator.GetNowData(city)
	if err != nil {
		log.Fatalf("Ошибка получения погоды: %v", err)
	}
	fmt.Println(weatherStr)
}
