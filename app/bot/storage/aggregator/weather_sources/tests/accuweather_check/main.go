package main

import (
	"fmt"
	"weather_bot/aggregator/weather_sources"
	"weather_bot/config"
)

func main() {
	city := "Владивосток"
	config.LoadAll()

	forecast, err := weather_sources.GetAccuWeatherForecast(city)
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	fmt.Println("Город:", forecast.City)
	fmt.Println("Утро:", forecast.Morning)
	fmt.Println("День:", forecast.Day)
	fmt.Println("Вечер:", forecast.Evening)
	fmt.Println("Ночь:", forecast.Night)

	now, err := weather_sources.GetAccuWeatherNow(city)
	if err != nil {
		fmt.Println("error", err)
		return
	}
	fmt.Println("Погода сейчас:", now)

	err = weather_sources.SaveRawAccuWeatherJSON(city)
	if err != nil {
		fmt.Println("Ошибка при сохранении JSON:", err)
	}
}
