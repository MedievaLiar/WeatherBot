package main

import (
	"app/bot/config"
	"app/bot/storage/aggregator/weather_sources"
	"fmt"
)

func main() {
	config.LoadAll()
	city := "Екатеринбург"

	forecastToday, err := weather_sources.GetYandexTodayForecast(city)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	fmt.Println("Город:", forecastToday.City)
	fmt.Println("Утро:", forecastToday.Morning)
	fmt.Println("День:", forecastToday.Day)
	fmt.Println("Вечер:", forecastToday.Evening)
	fmt.Println("Ночь:", forecastToday.Night)

	forecastTomorrow, err := weather_sources.GetYandexTomorrowForecast(city)
	fmt.Println("Город:", forecastTomorrow.City)
	fmt.Println("Утро:", forecastTomorrow.Morning)
	fmt.Println("День:", forecastTomorrow.Day)
	fmt.Println("Вечер:", forecastTomorrow.Evening)
	fmt.Println("Ночь:", forecastTomorrow.Night)

	forecastNow, err := weather_sources.GetYandexNow(city)
	if err != nil {
		fmt.Println("Ошибка", err)
		return
	}

	fmt.Println("Сейчас:", forecastNow)
}
