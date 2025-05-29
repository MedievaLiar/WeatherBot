package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"gopkg.in/yaml.v3"
	"sync"
)

type cachedData[T any] struct {
	LastUpdated time.Time `yaml:"last_updated"`
	Data        T         `yaml:"data"`
}

const (
	NowCacheDir      = "./bot/storage/now/now_cache"
	TodayCacheDir    = "./bot/storage/today/today_cache"
	TomorrowCacheDir = "./bot/storage/tomorrow/tomorrow_cache"

	NowTTL      = 30 * time.Minute
	TodayTTL    = 4 * time.Hour
	TomorrowTTL = 6 * time.Hour
)

func getCachePath(dir, city string) string {
	return filepath.Join(dir, fmt.Sprintf("%s.yaml", city))
}

func ReadCache[T any](dir, city string) (T, time.Time, error) {
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

func WriteCache[T any](dir string, city string, value T) error {
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

var cityLocks = make(map[string]*sync.Mutex)
var cityLocksMu sync.Mutex

func GetCityLock(city string) *sync.Mutex {
	cityLocksMu.Lock()
	defer cityLocksMu.Unlock()

	lock, ok := cityLocks[city]
	if !ok {
		lock = &sync.Mutex{}
		cityLocks[city] = lock
	}
	return lock
}

