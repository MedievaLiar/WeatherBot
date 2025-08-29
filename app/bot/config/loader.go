package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"gopkg.in/yaml.v3"
)

type City struct {
	Lat          float64 `yaml:"lat"`
	Lon          float64 `yaml:"lon"`
	SlugYandex   string  `yaml:"slug_yandex"`
	SlugGismeteo string  `yaml:"slug_gismeteo"`
	GismeteoID   string  `yaml:"gismeteo_number"`
	Timezone     string  `yaml:"timezone"`
}

type APIKeys struct {
	AccuWeather      string `yaml:"accuweather"`
	OpenWeather      string `yaml:"openweather"`
	TelegramBotToken string `yaml:"telegram_bot_token"`
}

var CityData map[string]City
var AccuLocationKeys map[string]string
var Keys APIKeys

func LoadAll() {
	loadYAML("cities_data.yaml", &CityData)
	loadYAML("accu_keys.yaml", &AccuLocationKeys)
	loadYAML("api_keys.yaml", &Keys)
}

func loadYAML(filename string, out any) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("Не удалось определить путь к файлу загрузчика")
	}

	dir := filepath.Dir(thisFile)
	fullPath := filepath.Join(dir, filename)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		log.Fatalf("Ошибка чтения %s: %v", fullPath, err)
	}
	if err := yaml.Unmarshal(data, out); err != nil {
		log.Fatalf("Ошибка разбора %s: %v", fullPath, err)
	}
}


