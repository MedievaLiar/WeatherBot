package weather_sources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"weather_bot/config"
)

func SaveRawOpenWeatherJSON(city string) error {
	apiKey := config.Keys.OpenWeather
	query := url.QueryEscape(city)
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=metric&lang=ru", query, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data OpenWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("aggregator/tests/openweather_check/raw_open/openweather_raw_%s.json", time.Now().Format("2006-01-02_15-04"))
	return os.WriteFile(filename, raw, 0644)
}

