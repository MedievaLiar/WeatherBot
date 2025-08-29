package storage

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"sync"
	"time"

	"app/bot/models"
	"app/bot/storage/aggregator"
	"app/bot/storage/aggregator/format"
)

type cachedData[T any] struct {
	LastUpdated time.Time `yaml:"last_updated"`
	Data        T         `yaml:"data"`
}

var (
	currentCities = make(map[int64]string)
	cityLocks     = make(map[string]*sync.Mutex)
	cityLocksMu   sync.Mutex
	cacheMutex    sync.RWMutex
)

const (
	NowCacheDir      = "./bot/storage/cache/now"
	TodayCacheDir    = "./bot/storage/cache/today"
	TomorrowCacheDir = "./bot/storage/cache/tomorrow"

	NowTTL      = 30 * time.Minute
	TodayTTL    = 4 * time.Hour
	TomorrowTTL = 6 * time.Hour
)

func GetNowForecast(city string) (string, error) {
	lock := getCityLock(city)
	lock.Lock()
	defer lock.Unlock()

	data, ts, err := readCache[models.PeriodWeather](NowCacheDir, city)
	if err != nil || time.Since(ts) > NowTTL {
		data, err = aggregator.GetNowData(city)
		if err != nil {
			return "", fmt.Errorf("ошибка получения погоды сейчас: %w", err)
		}
		if err := writeCache(NowCacheDir, city, data); err != nil {
			return "", fmt.Errorf("ошибка записи кэша: %w", err)
		}
	}

	return format.FormatNowWeather(city, data), nil
}

func GetTodayForecast(city string) (string, error) {
	lock := getCityLock(city)
	lock.Lock()
	defer lock.Unlock()

	data, ts, err := readCache[models.Forecast](TodayCacheDir, city)
	if err != nil || time.Since(ts) > TodayTTL {
		data, err = aggregator.GetTodayData(city)
		if err != nil {
			return "", fmt.Errorf("ошибка получения прогноза: %w", err)
		}
		if err := writeCache(TodayCacheDir, city, data); err != nil {
			return "", fmt.Errorf("ошибка записи кэша: %w", err)
		}
	}

	return format.FormatTodayWeather(data), nil
}

func GetTomorrowForecast(city string) (string, error) {
	lock := getCityLock(city)
	lock.Lock()
	defer lock.Unlock()

	data, ts, err := readCache[models.Forecast](TomorrowCacheDir, city)
	if err != nil || time.Since(ts) > TomorrowTTL {
		data, err = aggregator.GetTomorrowData(city)
		if err != nil {
			return "", fmt.Errorf("ошибка получения прогноза: %w", err)
		}
		if err := writeCache(TomorrowCacheDir, city, data); err != nil {
			return "", fmt.Errorf("ошибка записи кэша: %w", err)
		}
	}

	return format.FormatTomorrowWeather(data), nil
}

func GetCurrentCity(chatID int64) (string, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if city, ok := currentCities[chatID]; ok {
		return city, true
	}

	currentCity, ok := ExtractCurrentCity(chatID)
	if ok {
		currentCities[chatID] = currentCity
		return currentCity, true
	}
	return "", false
}

func SetCurrentCityInCache(chatID int64, city string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	currentCities[chatID] = city
}

func GetForecastCity(chatID int64) (string, bool) {
	forecastCity, ok := ExtractForecastCity(chatID)
	return forecastCity, ok && forecastCity != ""
}

func getCachePath(dir, city string) string {
	return filepath.Join(dir, fmt.Sprintf("%s.yaml", city))
}

func readCache[T any](dir, city string) (T, time.Time, error) {
	var wrapper cachedData[T]
	path := getCachePath(dir, city)

	data, err := os.ReadFile(path)
	if err != nil {
		var empty T
		return empty, time.Time{}, err
	}

	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		var empty T
		return empty, time.Time{}, err
	}

	return wrapper.Data, wrapper.LastUpdated, nil
}

func writeCache[T any](dir string, city string, value T) error {
	path := getCachePath(dir, city)
	tmp := path + ".tmp"

	wrapper := cachedData[T]{
		LastUpdated: time.Now(),
		Data:        value,
	}

	data, err := yaml.Marshal(&wrapper)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

func getCityLock(city string) *sync.Mutex {
	cityLocksMu.Lock()
	defer cityLocksMu.Unlock()

	lock, ok := cityLocks[city]
	if !ok {
		lock = &sync.Mutex{}
		cityLocks[city] = lock
	}
	return lock
}
