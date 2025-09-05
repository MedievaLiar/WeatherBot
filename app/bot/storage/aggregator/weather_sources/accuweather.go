package weather_sources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"app/bot/config"
	"app/bot/logger"
	"app/bot/models"
	"app/bot/storage/aggregator/weather_sources/utils"
)

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

type AccuWeatherResponse []AccuWeatherResponseElement

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

	var data AccuWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		logger.Error("AccuWeather: ошибка парсинга JSON для %s: %v", city, err)
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

	var data AccuWeatherResponse

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		logger.Error("AccuWeatherNow: ошибка парсинга JSON для %s: %v", city, err)
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
