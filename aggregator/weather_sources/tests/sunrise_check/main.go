package main

import (
	"fmt"
	"weather_bot/aggregator/weather_sources"
	"weather_bot/config"
)

func main() {
	city := "Екатеринбург"
	config.LoadAll()

	data, err := weather_sources.GetSunriseSunset(city, true)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	fmt.Printf("Время рассвета в городе %s, %s\n", city, data.Sunrise)
	fmt.Printf("Время заката в городе %s, %s\n", city, data.Sunset)
}
