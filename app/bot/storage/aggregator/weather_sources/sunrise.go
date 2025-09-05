package weather_sources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"app/bot/config"
	"app/bot/logger"
	"app/bot/models"
)

type sunriseSunsetResponse struct {
	Results struct {
		Sunrise string `json:"sunrise"`
		Sunset  string `json:"sunset"`
	} `json:"results"`
	Status string `json:"status"`
}

func GetSunriseSunset(city string, day bool) (models.Forecast, error) {
	info := config.CityData[city]
	loc, _ := time.LoadLocation(info.Timezone)

	// для  какой даты нужны данные
	var targetDate time.Time

	if day {
		targetDate = time.Now().In(loc)
	} else {
		targetDate = time.Now().In(loc).Add(24 * time.Hour)
	}
	dateStr := targetDate.Format("2006-01-02")

	url := fmt.Sprintf(
		"https://api.sunrise-sunset.org/json?lat=%f&lng=%f&date=%s&formatted=0",
		info.Lat, info.Lon, dateStr,
	)

	resp, err := http.Get(url)
	if err != nil {
		logger.Error("SunriseSunset: ошибка сети для %s: %v", city, err)
		return models.Forecast{}, err
	}
	defer resp.Body.Close()

	var data sunriseSunsetResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		logger.Error("SunriseSunset: ошибка парсинга JSON для %s: %v", city, err)
		return models.Forecast{}, err
	}

	if data.Status != "OK" {
		logger.Error("SunriseSunset: API ошибка для %s: %s", city, data.Status)
		return models.Forecast{}, fmt.Errorf("sunrise API error: %s", data.Status)
	}

	sunriseTime, err1 := time.Parse(time.RFC3339, data.Results.Sunrise)
	sunsetTime, err2 := time.Parse(time.RFC3339, data.Results.Sunset)

	if err1 != nil || err2 != nil {
		logger.Error(
			"SunriseSunset: неверный формат времени для %s: sunrise=%s sunset=%s",
			city, data.Results.Sunrise, data.Results.Sunset,
		)
		return models.Forecast{}, fmt.Errorf("invalid time format from API")
	}

	return models.Forecast{
		Sunrise: sunriseTime.In(loc).Format("15:04"),
		Sunset:  sunsetTime.In(loc).Format("15:04"),
	}, nil
}
