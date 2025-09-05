package weather_sources

import (
	"fmt"
	//"log"
	"net/http"
	"strings"
	"time"

	"app/bot/config"
	"app/bot/logger"
	"app/bot/models"
	"app/bot/storage/aggregator/weather_sources/utils"
	"github.com/PuerkitoBio/goquery"
)

func GetGismeteoNow(city string) (models.PeriodWeather, error) {
	info := config.CityData[city]
	url := fmt.Sprintf(
		"https://www.gismeteo.ru/weather-%s-%s/now/",
		info.SlugGismeteo, info.GismeteoID)
	//log.Println("🌐 Отправляем запрос к Gismeteo:", url)

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return models.PeriodWeather{}, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	// маскируемся под браузер
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return models.PeriodWeather{}, fmt.Errorf("ошибка при запросе: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return models.PeriodWeather{}, fmt.Errorf("ошибка запроса: %d", resp.StatusCode)
	}

	//log.Println("✅ Ответ получен, начинаем парсинг")

	// создаем goQuery документ для поиска по css селекторам
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return models.PeriodWeather{}, fmt.Errorf("ошибка чтения HTML: %w", err)
	}

	var weather models.PeriodWeather

	// извлечение температуры
	if temp := doc.Find(".now-weather temperature-value").First(); temp != nil {
		if val, exists := temp.Attr("value"); exists { // атрибут value
			weather.Temperature = float64(utils.ParseInt(val))
			//log.Println("🌡 Температура:", val)
		}
	}

	// по ощущениям
	if feel := doc.Find(".now-feel temperature-value").First(); feel != nil {
		if val, exists := feel.Attr("value"); exists {
			weather.FeelsLike = float64(utils.ParseInt(val))
			//log.Println("🌡 По ощущению:", val)
		}
	}

	// now-info: ветер и влажность
	foundWind := false
	foundHumidity := false

	doc.Find(".now-info-item").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".item-title").Text())

		switch title {
		case "Ветер":
			if !foundWind {
				if wind := s.Find("speed-value").First(); wind != nil {
					if val, exists := wind.Attr("value"); exists {
						weather.WindSpeed = utils.ParseFloat(val)
						//log.Println("💨 Ветер:", val)
						foundWind = true
					}
				}
			}
		case "Влажность":
			if !foundHumidity {
				value := s.Find(".item-value").First().Text()
				weather.Humidity = utils.ParseInt(value)
				//log.Println("💧 Влажность:", value)
				foundHumidity = true
			}
		}
	})

	if !validateWeather(weather, city) {
		logger.Error(
			"Gismeteo: получены невалидные данные для города %s", city)
		return models.PeriodWeather{}, fmt.Errorf("невалидные данные погоды")
	}
	return weather, nil
}

func validateWeather(w models.PeriodWeather, city string) bool {
	isValid := true

	if w.Temperature == 0 && w.FeelsLike == 0 && w.Humidity == 0 && w.WindSpeed == 0 {
		logger.Error("Gismeteo: все поля пустые для города %s", city)
		return false
	}

	if w.Temperature < -70 || w.Temperature > 60 {
		logger.Error("Gismeteo: неверная температура %0.1f для города %s", w.Temperature, city)
		isValid = false
	}

	if w.Humidity < 0 || w.Humidity > 100 {
		logger.Error("Gismeteo: неверная влажность %d для города %s", w.Humidity, city)
		isValid = false
	}

	if w.WindSpeed < 0 || w.WindSpeed > 200 {
		logger.Error("Gismeteo: неверная скорость ветра %0.1f для города %s", w.WindSpeed, city)
		isValid = false
	}

	return isValid
}
