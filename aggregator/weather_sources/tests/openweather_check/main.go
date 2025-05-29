package main

import (
	"fmt"
	"weather_bot/aggregator/weather_sources"
	"weather_bot/config"
)

func main() {
	config.LoadAll()

	city := "Владивосток"

	forecast, err := weather_sources.GetOpenWeatherForecast(city)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Println("Город:", forecast.City)
	fmt.Println("Утро:", forecast.Morning)
	fmt.Println("День:", forecast.Day)
	fmt.Println("Вечер:", forecast.Evening)
	fmt.Println("Ночь:", forecast.Night)

	/*err = weather_sources.SaveRawOpenWeatherJSON(city)
	if err != nil {
		fmt.Println("Ошибка при сохранении JSON:", err)
	}
	fmt.Println("JSON успешно сохранен.")*/
}
