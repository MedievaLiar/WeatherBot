package main

import (
	"fmt"
	"weather_bot/aggregator/weather_sources"
	"weather_bot/config"
)

func main() {
	config.LoadAll()
	city := "Уссурийск"

	/*forecast, err := weather_sources.GetYandexTodayForecast(city)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	fmt.Println("Город:", forecast.City)
	fmt.Println("Утро:", forecast.Morning)
	fmt.Println("День:", forecast.Day)
	fmt.Println("Вечер:", forecast.Evening)
	fmt.Println("Ночь:", forecast.Night)*/

	forecast, err := weather_sources.GetYandexNow(city)
	if err != nil {
		fmt.Println("Ошибка", err)
		return
	}

	fmt.Println(forecast)
}

