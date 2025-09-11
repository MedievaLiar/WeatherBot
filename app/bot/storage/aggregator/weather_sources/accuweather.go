package weather_sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"app/bot/config"
	"app/bot/logger"
	"app/bot/models"
	"app/bot/storage/aggregator/weather_sources/utils"
)

type AccuWeatherError struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type AccuWeatherResponse []AccuWeatherResponseElement

type AccuWeatherResponseElement struct {
	DateTime    string `json:"DateTime"`
	Temperature struct {
		Value float64 `json:"Value"`
	} `json:"Temperature"`
	FeelsLike struct {
		Value float64 `json:"Value"`
	} `json:"RealFeelTemperature"`
	Humidity int `json:"RelativeHumidity"`
	Wind     struct {
		Speed struct {
			Value float64 `json:"Value"`
		} `json:"Speed"`
	} `json:"Wind"`
	IconPhrase        string `json:"IconPhrase"`
	HasPrecipitation  bool   `json:"HasPrecipitation"`
	PrecipitationType string `json:"PrecipitationType"`
}

func GetAccuWeatherForecast(city string) (models.Forecast, error) {
	apiKey := config.Keys.AccuWeather
	locationKey := config.AccuLocationKeys[city]
	url := fmt.Sprintf(
		"http://dataservice.accuweather.com/forecasts/v1/hourly/12hour/%s?apikey=%s&language=ru&details=true&metric=true",
		locationKey, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		logger.Error("AccuWeather: ошибка сети для %s: %v", city, err)
		return models.Forecast{}, err
	}
	defer resp.Body.Close()

	// читаем весь ответ для проверки на ошибки
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("AccuWeather: ошибка чтения ответа для %s: %v", city, err)
		return models.Forecast{}, err
	}

	// проверяем, не является ли ответ ошибкой API
	var apiError struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal(bodyBytes, &apiError); err == nil && apiError.Code != "" {
		logger.Error("AccuWeather API error for %s: %s - %s", city, apiError.Code, apiError.Message)
		return models.Forecast{}, fmt.Errorf("API error: %s", apiError.Message)
	}

	// пытаемся парсить как массив (нормальный ответ)
	var data AccuWeatherResponse
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		logger.Error("AccuWeather: ошибка парсинга JSON для %s: %v. Response: %s", city, err, string(bodyBytes))
		return models.Forecast{}, err
	}

	if len(data) == 0 {
		logger.Error("AccuWeather: пустой ответ для города %s", city)
		return models.Forecast{}, fmt.Errorf("пустой ответ от API")
	}

	return processAccuWeatherData(data, city), nil
}

func processAccuWeatherData(data AccuWeatherResponse, city string) models.Forecast {
	loc, _ := time.LoadLocation(config.CityData[city].Timezone)

	// извлекаем данные из каждого элемента
	sample := func(elem AccuWeatherResponseElement) (time.Time, float64, float64, int, float64, bool) {
		dt, err := time.Parse(time.RFC3339, elem.DateTime)
		if err != nil {
			return time.Time{}, 0, 0, 0, 0, false
		}
		return dt, elem.Temperature.Value, elem.FeelsLike.Value, elem.Humidity, elem.Wind.Speed.Value, true
	}

	// передаем в универсальные обработчик
	return utils.ProcessWeatherData(data, city, sample, true, loc)
}

func GetAccuWeatherNow(city string) (models.PeriodWeather, error) {
	apiKey := config.Keys.AccuWeather
	locationKey := config.AccuLocationKeys[city]
	url := fmt.Sprintf(
		"http://dataservice.accuweather.com/forecasts/v1/hourly/1hour/%s?apikey=%s&details=true&metric=true",
		locationKey, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		logger.Error("AccuWeatherNow: ошибка сети для %s: %v", city, err)
		return models.PeriodWeather{}, err
	}
	defer resp.Body.Close()

	// Читаем весь ответ для проверки на ошибки
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("AccuWeatherNow: ошибка чтения ответа для %s: %v", city, err)
		return models.PeriodWeather{}, err
	}

	// Проверяем, не является ли ответ ошибкой API
	var apiError struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal(bodyBytes, &apiError); err == nil && apiError.Code != "" {
		logger.Error("AccuWeatherNow API error for %s: %s - %s", city, apiError.Code, apiError.Message)
		return models.PeriodWeather{}, fmt.Errorf("API error: %s", apiError.Message)
	}

	// Пытаемся парсить как массив (нормальный ответ)
	var data AccuWeatherResponse
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		logger.Error("AccuWeatherNow: ошибка парсинга JSON для %s: %v. Response: %s", city, err, string(bodyBytes))
		return models.PeriodWeather{}, err
	}

	if len(data) == 0 {
		logger.Error("AccuWeatherNow: пустой ответ для города %s", city)
		return models.PeriodWeather{}, fmt.Errorf("пустой ответ от AccuWeather")
	}

	elem := data[0]

	return models.PeriodWeather{
		Temperature: elem.Temperature.Value,
		FeelsLike:   elem.FeelsLike.Value,
		Humidity:    elem.Humidity,
		WindSpeed:   elem.Wind.Speed.Value / 3.6,
	}, nil
}
