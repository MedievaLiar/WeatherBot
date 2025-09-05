package weather_sources

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"app/bot/config"
	"app/bot/logger"
	"app/bot/models"
	"app/bot/storage/aggregator/weather_sources/utils"

	"github.com/PuerkitoBio/goquery"
)

func GetYandexTodayForecast(city string) (models.Forecast, error) {
	return getYandexForecast(city, true)
}

func GetYandexTomorrowForecast(city string) (models.Forecast, error) {
	return getYandexForecast(city, false)
}

func getYandexForecast(city string, isToday bool) (models.Forecast, error) {
	info := config.CityData[city]
	var url string
	var dayCard *goquery.Selection

	loc, err := time.LoadLocation(info.Timezone)
	targetDay := time.Now().In(loc)

	if !isToday {
		targetDay = targetDay.Add(24 * time.Hour)
		url = fmt.Sprintf("https://yandex.ru/pogoda/ru/%s/details/tomorrow", info.SlugYandex)
	} else {
		url = fmt.Sprintf("https://yandex.ru/pogoda/ru/%s/details/today", info.SlugYandex)
	}

	doc, err := getYandexDocument(url)
	if err != nil {
		logger.Error("Yandex: ошибка загрузки страницы для %s: %v", city, err)
		return models.Forecast{}, err
	}

	if !isToday {
		dayCard = doc.Find(fmt.Sprintf("article[data-day='%d_%d']", 2, targetDay.Day()))
	} else {
		dayCard = doc.Find(fmt.Sprintf("article[data-day='%d_%d']", 1, targetDay.Day()))
	}

	if dayCard.Length() == 0 {
		logger.Error("Yandex: не найден блок с data-day='%d_%d' для города %s", targetDay.Weekday(), targetDay.Day(), city)
		return models.Forecast{}, fmt.Errorf("блок дня не найден")
	}

	forecast := models.Forecast{
		City:    city,
		Morning: parsePeriod(dayCard, "m"),
		Day:     parsePeriod(dayCard, "d"),
		Evening: parsePeriod(dayCard, "e"),
		Night:   parsePeriod(dayCard, "n"),
	}

	return forecast, nil
}

func GetYandexNow(city string) (string, error) {
	info := config.CityData[city]
	url := fmt.Sprintf("https://yandex.ru/pogoda/%s", info.SlugYandex)

	doc, err := getYandexDocument(url)
	if err != nil {
		logger.Error("YandexNow: ошибка загружки для %s: %v", city, err)
		return "", err
	}

	desc := doc.Find("p.AppFact_warning__8kUUn").First().Text()
	if desc == "" {
		logger.Error("YandexNow: не найден блок с погодой для %s", city)
		return "", fmt.Errorf("блок погоды не найден")
	}

	result := strings.TrimSpace(desc)
	return result, nil
}

func parsePeriod(dayCard *goquery.Selection, prefix string) models.PeriodWeather {
	extract := func(styleValue, class string) string {
		return extractValueFromCard(dayCard, styleValue, class)
	}

	return models.PeriodWeather{
		Temperature:   utils.ParseTemperature(extract("grid-area:"+prefix+"-temp", "")),
		FeelsLike:     utils.ParseTemperature(extract("grid-area:"+prefix+"-feels", "")),
		Humidity:      utils.ParseInt(extract("grid-area:"+prefix+"-hum", "")),
		WindSpeed:     utils.ParseFloat(extract("grid-area:"+prefix+"-wind", "AppForecastDayPart_wind__k3V5t")),
		Precipitation: extract("grid-area:"+prefix+"-text", ""),
	}
}

func extractValueFromCard(card *goquery.Selection, styleValue, requiredClass string) string {
	var value string
	card.Find("div").Each(func(i int, s *goquery.Selection) {
		style, hasStyle := s.Attr("style")
		classAttr, hasClass := s.Attr("class")
		if hasStyle && strings.Contains(style, styleValue) && hasClass {
			if requiredClass == "" || strings.Contains(classAttr, requiredClass) {
				value = s.Text()
				return
			}
		}
	})

	if value == "" {
		logger.Error("Не найдено значение для styleValue=%s, class=%s", styleValue, requiredClass)
	}

	return value
}

func getYandexDocument(url string) (*goquery.Document, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ошибка запроса: %d", resp.StatusCode)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}
