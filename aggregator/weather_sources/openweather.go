package weather_sources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"weather_bot/aggregator/weather_sources/utils"
	"weather_bot/config"
	"weather_bot/models"
)

type OpenWeatherItem struct {
	Dt    int64 `json:"st"`
	Main  struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	DtTxt string `json:"dt_txt"`
}

type OpenWeatherResponse struct {
	List []OpenWeatherItem `json:"list"`
}

func processOpenWeatherData(data OpenWeatherResponse, city string) models.Forecast {
	loc, _ := time.LoadLocation(config.CityData[city].Timezone)

	sample := func(elem OpenWeatherItem) (time.Time, float64, float64, int, float64, bool) {
		dt, err := time.Parse("2006-01-02 15:04:05", elem.DtTxt)
		if err != nil {
			return time.Time{}, 0, 0, 0, 0, false
		}
		return dt, elem.Main.Temp, elem.Main.FeelsLike, elem.Main.Humidity, elem.Wind.Speed, true
	}

	return utils.ProcessWeatherData(data.List, city, sample, false, loc)
}

func GetOpenWeatherForecast(city string) (models.Forecast, error) {
	apiKey := config.Keys.OpenWeather
	query := url.QueryEscape(city)
	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=metric&lang=ru",
		query, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return models.Forecast{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return models.Forecast{}, fmt.Errorf("bad response from OpenWeather: %d", resp.StatusCode)
	}

	var data OpenWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return models.Forecast{}, err
	}
	return processOpenWeatherData(data, city), nil
}
