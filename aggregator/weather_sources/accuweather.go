package weather_sources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"weather_bot/aggregator/weather_sources/utils"
	"weather_bot/config"
	"weather_bot/models"
)

type AccuWeatherResponseElement struct {
	DateTime string `json:"DateTime"`
	Temperature struct {
		Value float64 `json:"Value"`
	} `json:"Temperature"`
	FeelsLike struct {
		Value float64 `json:"Value"`
	} `json:"RealFeelTemperature"`
	Humidity int `json:"RelativeHumidity"`
	Wind struct {
		Speed struct {
			Value float64 `json:"Value"`
		} `json:"Speed"`
	} `json:"Wind"`
	IconPhrase        string `json:"IconPhrase"`
	HasPrecipitation  bool   `json:"HasPrecipitation"`
	PrecipitationType string `json:"PrecipitationType"`
}

type AccuWeatherResponse []AccuWeatherResponseElement

func processAccuWeatherData(data AccuWeatherResponse, city string) models.Forecast {
	loc, _ := time.LoadLocation(config.CityData[city].Timezone)

	sample := func(elem AccuWeatherResponseElement) (time.Time, float64, float64, int, float64, bool) {
		dt, err := time.Parse(time.RFC3339, elem.DateTime)
		if err != nil {
			return time.Time{}, 0, 0, 0, 0, false
		}
		return dt, elem.Temperature.Value, elem.FeelsLike.Value, elem.Humidity, elem.Wind.Speed.Value, true
	}

	return utils.ProcessWeatherData(data, city, sample, true, loc)
}

func GetAccuWeatherForecast(city string) (models.Forecast, error) {
	apiKey := config.Keys.AccuWeather
	locationKey := config.AccuLocationKeys[city]
	url := fmt.Sprintf(
		"http://dataservice.accuweather.com/forecasts/v1/hourly/12hour/%s?apikey=%s&language=ru&details=true&metric=true",
		locationKey, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return models.Forecast{}, err
	}
	defer resp.Body.Close()

	var data AccuWeatherResponse

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return models.Forecast{}, err
	}

	return processAccuWeatherData(data, city), nil
}

func GetAccuWeatherNow(city string) (models.PeriodWeather, error) {
	apiKey := config.Keys.AccuWeather
	locationKey := config.AccuLocationKeys[city]
	url := fmt.Sprintf(
		"http://dataservice.accuweather.com/forecasts/v1/hourly/1hour/%s?apikey=%s&details=true&metric=true",
		locationKey, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return models.PeriodWeather{}, err
	}
	defer resp.Body.Close()

	var data AccuWeatherResponse

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return models.PeriodWeather{}, err
	}

	if len(data) == 0 {
		return models.PeriodWeather{}, fmt.Errorf("пустой ответ от AccuWeather")
	}

	elem := data[0]

	return models.PeriodWeather{
		Temperature: elem.Temperature.Value,
		FeelsLike:	 elem.FeelsLike.Value,
		Humidity: 	 elem.Humidity,
		WindSpeed: 	 elem.Wind.Speed.Value / 3.6,
	}, nil
}

