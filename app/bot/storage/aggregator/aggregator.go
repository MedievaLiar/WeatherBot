package aggregator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"app/bot/logger"
	"app/bot/models"
	"app/bot/storage/aggregator/weather_sources"
)

func GetNowData(city string) (models.PeriodWeather, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	sources := make([]models.PeriodWeather, 0, 2)
	var yandexPrecip string

	wg.Add(2)

	// AccuWeather откатил бесплатный тариф, пока остается на всякий
	/*go func() {
		defer wg.Done()
		if accuNow, err := weather_sources.GetAccuWeatherNow(city); err == nil {
			mu.Lock()
			sources = append(sources, accuNow)
			mu.Unlock()
		} else {
			logger.Error("AccuWeatherNow: ошибка для %s: %v", city, err)
		}
	}()*/

	go func() {
		defer wg.Done()
		if gismeteoNow, err := weather_sources.GetGismeteoNow(city); err == nil {
			mu.Lock()
			sources = append(sources, gismeteoNow)
			mu.Unlock()
		} else {
			logger.Error("GismeteoNow: ошибка для %s: %v", city, err)
		}
	}()

	go func() {
		defer wg.Done()
		if precip, err := weather_sources.GetYandexNow(city); err == nil {
			mu.Lock()
			yandexPrecip = precip
			mu.Unlock()
		} else {
			logger.Error("YandexNow: ошибка для %s: %v", city, err)
		}
	}()

	wg.Wait()

	result := averageNowForecast(sources...)
	result.Precipitation = yandexPrecip

	if len(sources) == 0 {
		logger.Error("GetNowData: нет данных ни от одного источника для %s", city)
	}

	return result, nil
}

func GetTodayData(city string) (models.Forecast, error) {
	forecasts, err := getAllForecasts(city)
	if err != nil {
		return models.Forecast{}, fmt.Errorf("ошибка при получении прогнозов: %w", err)
	}

	if len(forecasts) == 0 {
		return models.Forecast{}, fmt.Errorf("не получено ни одного прогноза")
	}

	result := averageTodayForecast(forecasts)
	return result, nil
}

func GetTomorrowData(city string) (models.Forecast, error) {
	tomorrowForecast, err := weather_sources.GetYandexTomorrowForecast(city)
	if err != nil {
		logger.Error("GetTomorrowData: ошибка для %s: %v", city, err)
		return models.Forecast{}, fmt.Errorf("не удалось получить прогноз от Яндекса: %w", err)
	}

	sunData, err := weather_sources.GetSunriseSunset(city, false)
	if err == nil {
		tomorrowForecast.Sunrise = sunData.Sunrise
		tomorrowForecast.Sunset = sunData.Sunset
	}

	return tomorrowForecast, nil
}

func averageNowForecast(sources ...models.PeriodWeather) models.PeriodWeather {
	var result models.PeriodWeather
	var count float64

	for _, src := range sources {
		if src.Temperature == 0 && src.FeelsLike == 0 && src.Humidity == 0 && src.WindSpeed == 0 {
			continue
		}
		result.Temperature += src.Temperature
		result.FeelsLike += src.FeelsLike
		result.Humidity += src.Humidity
		result.WindSpeed += src.WindSpeed

		if result.Precipitation == "" && src.Precipitation != "" {
			result.Precipitation = src.Precipitation
		}

		count++
	}

	if count == 0 {
		return models.PeriodWeather{}
	}

	result.Temperature /= count
	result.FeelsLike /= count
	result.Humidity /= int(count)
	result.WindSpeed /= count

	return result
}

func averageTodayForecast(forecasts []models.Forecast) models.Forecast {
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

			temp += p.Temperature
			feels += p.FeelsLike
			humid += p.Humidity
			wind += p.WindSpeed

			if precip == "" && p.Precipitation != "" {
				precip = p.Precipitation
			}
			count++
		}

		if count == 0 {
			return models.PeriodWeather{}
		}

		return models.PeriodWeather{
			Temperature:   temp / float64(count),
			FeelsLike:     feels / float64(count),
			Humidity:      humid / count,
			WindSpeed:     wind / float64(count),
			Precipitation: precip,
		}
	}

	result.Morning = avgPeriod(func(f models.Forecast) models.PeriodWeather { return f.Morning })
	result.Day = avgPeriod(func(f models.Forecast) models.PeriodWeather { return f.Day })
	result.Evening = avgPeriod(func(f models.Forecast) models.PeriodWeather { return f.Evening })
	result.Night = avgPeriod(func(f models.Forecast) models.PeriodWeather { return f.Night })

	return result
}

func getAllForecasts(city string) ([]models.Forecast, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type result struct {
		forecast models.Forecast
		err      error
		source   string
	}

	resultsCh := make(chan result, 3)
	sources := map[string]func(string) (models.Forecast, error){
		"yandex":      weather_sources.GetYandexTodayForecast,
		"openweather": weather_sources.GetOpenWeatherForecast,
		//"accuweather": weather_sources.GetAccuWeatherForecast,
	}

	for name, source := range sources {
		go func(name string, f func(string) (models.Forecast, error)) {
			forecast, err := f(city)
			select {
			case resultsCh <- result{forecast, err, name}:
			case <-ctx.Done():
			}
		}(name, source)
	}

	forecasts := make([]models.Forecast, 0, 3)

	for i := 0; i < len(sources); i++ {
		select {
		case res := <-resultsCh:
			if res.err == nil {
				forecasts = append(forecasts, res.forecast)
			} else {
				log.Printf("Ошибка от %s: %v", res.source, res.err)
			}
		case <-ctx.Done():
			log.Printf("Таймаут получения прогнозов")
		}
	}

	// Добавляем данные о восходе/закате
	if len(forecasts) > 0 {
		sunData, err := weather_sources.GetSunriseSunset(city, true)
		if err == nil {
			forecasts[0].Sunrise = sunData.Sunrise
			forecasts[0].Sunset = sunData.Sunset
		}
	}

	return forecasts, nil
}
