package weather_sources

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	//"log"

	"app/bot/config"
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

	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return models.PeriodWeather{}, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return models.PeriodWeather{}, fmt.Errorf("ошибка при запросе: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return models.PeriodWeather{}, fmt.Errorf("ошибка запроса: %d", res.StatusCode)
	}

	//log.Println("✅ Ответ получен, начинаем парсинг")

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return models.PeriodWeather{}, fmt.Errorf("ошибка чтения HTML: %w", err)
	}

	var weather models.PeriodWeather

	// Температура
	if temp := doc.Find(".now-weather temperature-value").First(); temp != nil {
		if val, exists := temp.Attr("value"); exists {
			weather.Temperature = float64(utils.ParseInt(val))
			//log.Println("🌡 Температура:", val)
		}
	}

	// По ощущениям
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
	return weather, nil
}
