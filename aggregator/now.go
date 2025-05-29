package aggregator

import (
	"sync"
	"weather_bot/aggregator/weather_sources"
	"weather_bot/models"
)

func GetWeatherNow(city string) (models.PeriodWeather, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	sources := make([]models.PeriodWeather, 0, 2)
	var yandexPrecip string

	wg.Add(3)

	go func() {
		defer wg.Done()
		if accuNow, err := weather_sources.GetAccuWeatherNow(city); err == nil {
			mu.Lock()
			sources = append(sources, accuNow)
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if gismeteoNow, err := weather_sources.GetGismeteoNow(city); err == nil {
			mu.Lock()
			sources = append(sources, gismeteoNow)
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		if precip, err := weather_sources.GetYandexNow(city); err == nil {
			mu.Lock()
			yandexPrecip = precip
			mu.Unlock()
		}
	}()

	wg.Wait()

	result := averageNowWeather(sources...)
	result.Precipitation = yandexPrecip
	return result, nil
}


func averageNowWeather(sources ...models.PeriodWeather) models.PeriodWeather {
    var result models.PeriodWeather
    var count float64

    for _, src := range sources {
        if src.Temperature == 0 && src.FeelsLike == 0 && src.Humidity == 0 && src.WindSpeed == 0 {
            continue
        }
        result.Temperature += src.Temperature
        result.FeelsLike   += src.FeelsLike
        result.Humidity    += src.Humidity
        result.WindSpeed   += src.WindSpeed

		if result.Precipitation == "" && src.Precipitation != "" {
			result.Precipitation = src.Precipitation
		}

        count++
    }

    if count == 0 {
        return models.PeriodWeather{}
    }

    result.Temperature /= count
    result.FeelsLike   /= count
    result.Humidity    /= int(count)
    result.WindSpeed   /= count

    return result
}

