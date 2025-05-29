package aggregator

import (
	"fmt"
	"weather_bot/models"
	"weather_bot/aggregator/weather_sources"
)

func GetTomorrowForecast(city string) (models.Forecast, error) {
	tomorrowForecast, err := weather_sources.GetYandexTomorrowForecast(city)
	if err != nil {
		return models.Forecast{}, fmt.Errorf("не удалось получить прогноз от Яндекса: %w", err)
	}

	sunData, err := weather_sources.GetSunriseSunset(city, false)
	if err == nil {
		tomorrowForecast.Sunrise = sunData.Sunrise
		tomorrowForecast.Sunset  = sunData.Sunset
	}

	return tomorrowForecast, nil
}

