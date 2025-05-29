package weather_sources

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"github.com/PuerkitoBio/goquery"

	"weather_bot/models"
	"weather_bot/config"
	"weather_bot/aggregator/weather_sources/utils"
)

func getYandexForecast(city string, isToday bool) (models.Forecast, error) {
	info := config.CityData[city]
	url := fmt.Sprintf("https://yandex.ru/pogoda/ru/%s/details/today", info.SlugYandex)

	resp, err := http.Get(url)
	if err != nil {
		return models.Forecast{}, fmt.Errorf("ошибка при получении страницы: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return models.Forecast{}, fmt.Errorf("ошибка при парсинге страницы: %v", err)
	}

	loc, err := time.LoadLocation(info.Timezone)
	if err != nil {
		return models.Forecast{}, fmt.Errorf("ошибка при загрузке временной зоны: %v", err)
	}

	targetDay := time.Now().In(loc)
	if !isToday {
		targetDay = targetDay.Add(24 * time.Hour)
	}
	dayID := fmt.Sprintf("d_%d", targetDay.Day())

	dayCard := doc.Find(fmt.Sprintf("div[data-id='%s']", dayID))
	if dayCard.Length() == 0 {
		return models.Forecast{}, fmt.Errorf("не найден блок с data-id='%s'", dayID)
	}

	extractValue := func(styleValue string, requiredClass string) string {
		val := ""
		dayCard.Find("div").Each(func(i int, s *goquery.Selection) {
			style, hasStyle := s.Attr("style")
			classAttr, hasClass := s.Attr("class")
			if hasStyle && strings.Contains(style, styleValue) && hasClass {
				if strings.Contains(classAttr, requiredClass) {
					val = s.Text()
					return
				}
			}
		})
		return val
	}

	getPeriod := func(prefix string) models.PeriodWeather {
		return models.PeriodWeather{
			Temperature:   utils.ParseTemperature(extractValue("grid-area:"+prefix+"-temp", "")),
			FeelsLike:     utils.ParseTemperature(extractValue("grid-area:"+prefix+"-feels", "")),
			Humidity:      utils.ParseInt(extractValue("grid-area:"+prefix+"-hum", "")),
			WindSpeed:     utils.ParseFloat(extractValue("grid-area:"+prefix+"-wind", "AppForecastDayPart_wind__k3V5t")),
			Precipitation: extractValue("grid-area:"+prefix+"-text", ""),
		}
	}

	return models.Forecast{
		City:    city,
		Morning: getPeriod("m"),
		Day:     getPeriod("d"),
		Evening: getPeriod("e"),
		Night:   getPeriod("n"),
	}, nil
}

func GetYandexTodayForecast(city string) (models.Forecast, error) {
	return getYandexForecast(city, true)
}

func GetYandexTomorrowForecast(city string) (models.Forecast, error) {
	return getYandexForecast(city, false)
}

func GetYandexNow(city string) (string, error) {
	info := config.CityData[city]
	url := fmt.Sprintf("https://yandex.ru/pogoda/%s", info.SlugYandex)

	client := &http.Client{Timeout: 15 * time.Second} // увеличили таймаут
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка при запросе: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ошибка запроса: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения HTML: %w", err)
	}

	desc := doc.Find("p.AppFact_warning__8kUUn").First().Text()
	if desc == "" {
		return "", fmt.Errorf("не найден блок с погодой")
	}

	return strings.TrimSpace(desc), nil
}

