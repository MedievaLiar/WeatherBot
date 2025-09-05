package main

import (
	"app/bot/config"
	"app/bot/storage/aggregator/weather_sources"
	"fmt"
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
