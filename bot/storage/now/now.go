package now

import (
	"fmt"
	"time"

	"weather_bot/aggregator"
	"weather_bot/aggregator/format"
	"weather_bot/models"
	"weather_bot/bot/storage"
)

func GetWeatherNow(city string) (string, error) {
	lock := storage.GetCityLock(city)
	lock.Lock()
	defer lock.Unlock()

	data, ts, err := storage.ReadCache[models.PeriodWeather](storage.NowCacheDir, city)
	if err != nil || time.Since(ts) > storage.NowTTL {
		data, err = aggregator.GetWeatherNow(city)
		if err != nil {
			return "", fmt.Errorf("ошибка получения погоды сейчас: %w", err)
		}
		if err := storage.WriteCache(storage.NowCacheDir, city, data); err != nil {
			return "", fmt.Errorf("ошибка записи кэша: %w", err)
		}
	}

	return format.FormatNowWeather(city, data), nil
}

