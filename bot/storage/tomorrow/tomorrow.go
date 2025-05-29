package tomorrow

import (
	"fmt"
	"time"

	"weather_bot/aggregator"
	"weather_bot/aggregator/format"
	"weather_bot/models"
	"weather_bot/bot/storage"
)

func GetTomorrowForecast(city string) (string, error) {
	lock := storage.GetCityLock(city)
	lock.Lock()
	defer lock.Unlock()

	data, ts, err := storage.ReadCache[models.Forecast](storage.TomorrowCacheDir, city)
	if err != nil || time.Since(ts) > storage.TomorrowTTL {
		data, err = aggregator.GetTomorrowForecast(city)
		if err != nil {
			return "", fmt.Errorf("ошибка получения прогноза: %w", err)
		}
		if err := storage.WriteCache(storage.TomorrowCacheDir, city, data); err != nil {
			return "", fmt.Errorf("ошибка записи кэша: %w", err)
		}
	}

	return format.FormatTomorrowWeather(data), nil
}


