package main

import (
	"fmt"
	"weather_bot/aggregator/weather_sources"
	"weather_bot/config"
)

func main() {
	city := "Владивосток"
	config.LoadAll()

	now, err := weather_sources.GetGismeteoNow(city)
	if err != nil {
		fmt.Println("error", err)
		return
	}
	fmt.Println("Город:", city)
	fmt.Println("Погода сейчас:", now)
}
