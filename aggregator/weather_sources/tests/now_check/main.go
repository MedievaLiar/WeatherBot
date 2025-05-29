package main

import (
	"fmt"
	"log"
	"weather_bot/aggregator"
	"weather_bot/config"
)

func main() {
    city := "Москва"
	config.LoadAll()

    weatherStr, err := aggregator.GetWeatherNow(city)
    if err != nil {
        log.Fatalf("Ошибка получения погоды: %v", err)
    }
    fmt.Println(weatherStr)
}

