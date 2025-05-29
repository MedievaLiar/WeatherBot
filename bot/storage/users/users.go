package users

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"weather_bot/config"

	"maps"

	"gopkg.in/yaml.v3"
)

const (
	userCacheFile    = "./bot/storage/users/users.yaml"
	cacheSaveRetries = 3
)

type UserPreferences struct {
	ForecastCity         string `yaml:"forecast_city"`
	WantDailyForecast    bool   `yaml:"want_daily_forecast"`
	ForecastMskHour      int    `yaml:"forecast_hour"`
	ForecastLocalHour    int    `yaml:"forecast_local_hour"`
	AwaitingForecastHour bool   `yaml:"awaiting_forecast_hour"`
	WantChangeCity 		 bool   `yaml:"want_change_city"`
	LastCity			 string `yaml:"last_city,omitempty"`
}

var (
	userPrefs     = make(map[int64]UserPreferences)
	currentCities = make(map[int64]string)
	cacheMutex    sync.RWMutex
)

func SetCurrentCity(chatID int64, city string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	currentCities[chatID] = city

    if _, ok := userPrefs[chatID]; !ok {
        userPrefs[chatID] = UserPreferences{}
    }

	prefs := userPrefs[chatID]
    prefs.LastCity = city
    userPrefs[chatID] = prefs
	go safeSave()
}

func GetCurrentCity(chatID int64) (string, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if city, ok := currentCities[chatID]; ok {
        return city, true
    }

    if prefs, ok := userPrefs[chatID]; ok && prefs.LastCity != "" {
        currentCities[chatID] = prefs.LastCity
        return prefs.LastCity, true
    }
    return "", false
}

func SetForecastCity(chatID int64, city string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	prefs := userPrefs[chatID]
	prefs.ForecastCity = city
	userPrefs[chatID] = prefs
	go safeSave()
}

func GetForecastCity(chatID int64) (string, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	prefs, ok := userPrefs[chatID]
	return prefs.ForecastCity, ok && prefs.ForecastCity != ""
}

func SetUserForecast(chatID int64, wantForecast bool, mskHour int, localHour int) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	prefs := userPrefs[chatID]
	prefs.WantDailyForecast = wantForecast
	prefs.ForecastMskHour   = mskHour
	prefs.ForecastLocalHour = localHour
	userPrefs[chatID]       = prefs
	go safeSave()
}

func GetUserForecastPrefs(chatID int64) (bool, int, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	prefs, ok := userPrefs[chatID]
	if !ok {
		return false, 0, false
	}
	return prefs.WantDailyForecast, prefs.ForecastLocalHour, true
}

func SetAwaitingForecastHour(chatID int64, awaiting bool) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	prefs := userPrefs[chatID]
	prefs.AwaitingForecastHour = awaiting
	userPrefs[chatID] = prefs
	go safeSave()
}

func IsAwaitingForecastHour(chatID int64) bool {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return userPrefs[chatID].AwaitingForecastHour
}

func WantChangeDailyCity(chatID int64, want bool) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	prefs := userPrefs[chatID]
	prefs.WantChangeCity = want
	userPrefs[chatID] = prefs
	go safeSave()
}

func IsChangingCity(chatID int64) bool {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return userPrefs[chatID].WantChangeCity
}

func LoadUserCache() error {
	data, err := os.ReadFile(userCacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("[UserCache] Файл кэша не найден, будет создан новый")
			return nil
		}
		return fmt.Errorf("ошибка чтения файла кэша: %w", err)
	}

	var cached map[int64]UserPreferences
	if err := yaml.Unmarshal(data, &cached); err != nil {
		return fmt.Errorf("ошибка разбора YAML: %w", err)
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	maps.Copy(userPrefs, cached)

    for chatID, prefs := range cached {
        if prefs.LastCity != "" {
            currentCities[chatID] = prefs.LastCity
        }
    }

    fmt.Printf("[UserCache] Загружено %d пользователей (восстановлено %d текущих городов)\n",
        len(cached),
        len(currentCities))
    return nil
}

func saveUserCache() error {
	data, err := func() ([]byte, error) {
		cacheMutex.RLock()
		defer cacheMutex.RUnlock()
		return yaml.Marshal(userPrefs)
	}()

	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(userCacheFile), 0755); err != nil {
		return fmt.Errorf("ошибка создания директории: %w", err)
	}

	tmpFile := userCacheFile + ".tmp"
	for range cacheSaveRetries {
		if err := os.WriteFile(tmpFile, data, 0644); err == nil {
			if err := os.Rename(tmpFile, userCacheFile); err == nil {
				fmt.Printf("[UserCache] Кэш сохранён (%d пользователей)\n", len(userPrefs))
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("не удалось сохранить кэш после %d попыток", cacheSaveRetries)
}

func safeSave() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("[UserCache] Восстановлено после паники при сохранении:", r)
		}
	}()
	if err := saveUserCache(); err != nil {
		fmt.Println("[UserCache] Ошибка сохранения:", err)
	}
}

func StartAutoSave() {
	go func() {
		for {
			next := nextSaveTime(time.Now())
			time.Sleep(time.Until(next))
			safeSave()
		}
	}()
}

func nextSaveTime(now time.Time) time.Time {
	saveHours := []int{8, 12, 18, 22}
	for _, h := range saveHours {
		t := time.Date(now.Year(), now.Month(), now.Day(), h, 0, 0, 0, now.Location())
		if now.Before(t) {
			return t
		}
	}
	return time.Date(now.Year(), now.Month(), now.Day()+1, saveHours[0], 0, 0, 0, now.Location())
}

func GetAllUserPrefs() map[int64]UserPreferences {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return maps.Clone(userPrefs)
}

func ConvertToMsk(city string, localHour int) int {
	loc, _ := time.LoadLocation(config.CityData[city].Timezone)

	localTime := time.Date(2006, 1, 2, localHour, 0, 0, 0, loc)

	mskLoc, _ := time.LoadLocation("Europe/Moscow")
	mskTime := localTime.In(mskLoc)
	return mskTime.Hour()
}
