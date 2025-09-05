package main

import (
	"app/bot/config"
	"app/bot/storage/aggregator"
	"app/bot/storage/aggregator/format"
	"fmt"
)

func main() {
	city := "Владивосток"
	config.LoadAll()

	forecast, err := aggregator.GetTodayData(city)
	if err != nil {
		fmt.Printf("❌ Ошибка при получении прогноза: %v\n", err)
		return
	}

	message := format.FormatTodayWeather(forecast)
	fmt.Println(message)
}
