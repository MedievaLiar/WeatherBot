package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type City struct {
	Lat          float64 `yaml:"lat"`             // широта
	Lon          float64 `yaml:"lon"`             // долгота
	SlugYandex   string  `yaml:"slug_yandex"`     // как город забит в ссылке яндекса
	SlugGismeteo string  `yaml:"slug_gismeteo"`   // в ссылке гисметео
	GismeteoID   string  `yaml:"gismeteo_number"` // к ссылке добавляется числовой код
	Timezone     string  `yaml:"timezone"`
}

type APIKeys struct {
	AccuWeather      string
	OpenWeather      string
	TelegramBotToken string
	TelegramAdminID  string
}

var CityData map[string]City
var AccuLocationKeys map[string]string
var Keys APIKeys

// предзагрузка ключей для работы бота
func LoadAll() {
	loadEnv("api.env")                            // загружаем переменные окружения из файла
	loadYAML("cities_data.yaml", &CityData)       // загружаем данные городов
	loadYAML("accu_keys.yaml", &AccuLocationKeys) // загружаем ключи AccuWeather

	// заполняет структуру ключей из переменных окружения
	Keys = APIKeys{
		AccuWeather:      os.Getenv("ACCUWEATHER_API"),
		OpenWeather:      os.Getenv("OPENWEATHER_API"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_API"),
		TelegramAdminID:  os.Getenv("TELEGRAM_ADMIN_ID"),
	}

	checkRequiredEnvVars() // проверяем что все ключи подгружены
}

func loadYAML(filename string, out any) {
	data, err := readConfigFile(filename) // читаем файл
	if err != nil {
		log.Fatalf("Ошибка чтения %s: %v", filename, err)
	}

	if err := yaml.Unmarshal(data, out); err != nil { // парсим YAML
		log.Fatalf("Ошибка разбора %s: %v", filename, err)
	}

}

func loadEnv(filename string) {
	_, thisFile, _, ok := runtime.Caller(0) // путь к текущему файлу
	if !ok {
		log.Fatalf("Не удалось определить путь к файлу загрузчика")
	}

	dir := filepath.Dir(thisFile)
	envPath := filepath.Join(dir, "env", filename)

	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Не удалось загрузить %s: %v", envPath, err)
	}
}

func checkRequiredEnvVars() {
	required := map[string]string{
		"TELEGRAM_BOT_API":  Keys.TelegramBotToken,
		"TELEGRAM_ADMIN_ID": Keys.TelegramAdminID,
		"ACCUWEATHER_API":   Keys.AccuWeather,
		"OPENWEATHER_API":   Keys.OpenWeather,
	}

	for name, value := range required {
		if value == "" {
			log.Fatalf("Обязательная переменная окружения %s не установлена", name)
		}
	}
}

func readConfigFile(filename string) ([]byte, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("Не удалось определить путь к файлу загрузчика")
	}

	dir := filepath.Dir(thisFile)
	fullPath := filepath.Join(dir, filename)

	// читаем в память
	return os.ReadFile(fullPath)
}
