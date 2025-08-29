package main

import (
	"fmt"
	"weather_bot/aggregator"
	"weather_bot/aggregator/format"
	"weather_bot/config"
)

func main() {
	city := "Владивосток"
	config.LoadAll()

	forecast, err := aggregator.GetFinalForecast(city)
	if err != nil {
		fmt.Printf("❌ Ошибка при получении прогноза: %v\n", err)
		return
	}

	message := format.FormatTodayWeather(forecast)
	fmt.Println(message)
}

