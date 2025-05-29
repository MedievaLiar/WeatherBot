package aggregator

import (
	"fmt"
	"sync"
	"weather_bot/models"
	"weather_bot/aggregator/weather_sources"
)

func getAllForecasts(city string) ([]models.Forecast, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	forecasts := make([]models.Forecast, 0, 3)

	wg.Add(3)

	go func() {
		defer wg.Done()
		if yandexForecast, err := weather_sources.GetYandexTodayForecast(city); err == nil {
			mu.Lock()
			forecasts = append(forecasts, yandexForecast)
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if openWeatherForecast, err := weather_sources.GetOpenWeatherForecast(city); err == nil {
			mu.Lock()
			forecasts = append(forecasts, openWeatherForecast)
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if accuWeatherForecast, err := weather_sources.GetAccuWeatherForecast(city); err == nil {
			mu.Lock()
			forecasts = append(forecasts, accuWeatherForecast)
			mu.Unlock()
		}
	}()

	wg.Wait()

	sunData, err := weather_sources.GetSunriseSunset(city, true)
	if err == nil && len(forecasts) > 0 {
		forecasts[0].Sunrise = sunData.Sunrise
		forecasts[0].Sunset = sunData.Sunset
	}

	return forecasts, nil
}

func GetFinalForecast(city string) (models.Forecast, error) {
	forecasts, err := getAllForecasts(city)
	if err != nil {
		return models.Forecast{}, fmt.Errorf("ошибка при получении прогнозов: %w", err)
	}

	if len(forecasts) == 0 {
		return models.Forecast{}, fmt.Errorf("не получено ни одного прогноза")
	}

	return averageForecast(forecasts), nil
}

func averageForecast(forecasts []models.Forecast) models.Forecast {
	if len(forecasts) == 0 {
		return models.Forecast{}
	}

	result := models.Forecast{
		City:    forecasts[0].City,
		Sunrise: forecasts[0].Sunrise,
		Sunset:  forecasts[0].Sunset,
	}

	avgPeriod := func(getValue func(models.Forecast) models.PeriodWeather) models.PeriodWeather {
		var wind, feels, temp float64
		var humid, count int
		precip := ""

		for _, f := range forecasts {
			p := getValue(f)

			if p.Temperature == 0 && p.FeelsLike == 0 && p.Humidity == 0 && p.WindSpeed == 0 {
				continue
			}

			temp  += p.Temperature
			feels += p.FeelsLike
			humid += p.Humidity
			wind  += p.WindSpeed
			if precip == "" && p.Precipitation != "" {
				precip = p.Precipitation
			}
			count++
		}

		if count == 0 {
			return models.PeriodWeather{}
		}

		return models.PeriodWeather{
			Temperature:   temp  / float64(count),
			FeelsLike:     feels / float64(count),
			Humidity:      humid / count,
			WindSpeed:     wind  / float64(count),
			Precipitation: precip,
		}
	}

	result.Morning = avgPeriod(func(f models.Forecast) models.PeriodWeather { return f.Morning 	})
	result.Day     = avgPeriod(func(f models.Forecast) models.PeriodWeather { return f.Day 		})
	result.Evening = avgPeriod(func(f models.Forecast) models.PeriodWeather { return f.Evening 	})
	result.Night   = avgPeriod(func(f models.Forecast) models.PeriodWeather { return f.Night 	})

	return result
}

