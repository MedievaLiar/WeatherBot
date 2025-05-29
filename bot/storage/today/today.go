package today

import (
	"fmt"
	"time"

	"weather_bot/aggregator"
	"weather_bot/aggregator/format"
	"weather_bot/models"
	"weather_bot/bot/storage"
)

func GetTodayForecast(city string) (string, error) {
	lock := storage.GetCityLock(city)
	lock.Lock()
	defer lock.Unlock()

	data, ts, err := storage.ReadCache[models.Forecast](storage.TodayCacheDir, city)
	if err != nil || time.Since(ts) > storage.TodayTTL {
		data, err = aggregator.GetFinalForecast(city)
		if err != nil {
			return "", fmt.Errorf("ошибка получения прогноза: %w", err)
		}
		if err := storage.WriteCache(storage.TodayCacheDir, city, data); err != nil {
			return "", fmt.Errorf("ошибка записи кэша: %w", err)
		}
	}

	return format.FormatTodayWeather(data), nil
}

