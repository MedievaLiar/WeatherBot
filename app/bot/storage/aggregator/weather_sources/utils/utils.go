package utils

import (
	"strconv"
	"strings"
	"time"

	"app/bot/models"
)

// убираем лишнее для температуры
func ParseTemperature(s string) float64 {
	s = strings.ReplaceAll(s, "−", "-")
	s = strings.TrimLeft(s, "+")
	s = strings.TrimSpace(strings.TrimSuffix(s, "°"))
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

func ParseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return 0
	}

	val, _ := strconv.ParseFloat(parts[0], 64)
	return val
}

func ParseInt(s string) int {
	s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
	val, _ := strconv.Atoi(s)
	return val
}

// агрегация данных
type GenericSumData struct {
	Temp        float64
	FeelsLike   float64
	Humidity    int
	Wind        float64
	Count       int       // количество samples
	TempSamples []float64 // все значения температур для дебага
}

// накапливаем данные
func (s *GenericSumData) AddSample(temp, feelsLike float64, humidity int, wind float64) {
	s.Temp += temp
	s.FeelsLike += feelsLike
	s.Humidity += humidity
	s.Wind += wind
	s.TempSamples = append(s.TempSamples, temp) // сохраняем raw data
	s.Count++
}

// усредняем и возвращаем готовую структуруЖ
func (s *GenericSumData) Average(divideWind bool) models.PeriodWeather {
	if s.Count == 0 {
		return models.PeriodWeather{}
	}

	wind := s.Wind / float64(s.Count)
	if divideWind {
		wind = wind / 3.6 // конвертация км/ч -> м/с
	}

	return models.PeriodWeather{
		Temperature: s.Temp / float64(s.Count),
		FeelsLike:   s.FeelsLike / float64(s.Count),
		Humidity:    s.Humidity / s.Count,
		WindSpeed:   s.Wind / float64(s.Count) / 3.6,
	}
}

type SampleFunc[T any] func(elem T) (
	timestamp time.Time,
	temp,
	feelsLike float64,
	humidity int,
	wind float64,
	ok bool,
)

func ProcessWeatherData[T any](
	data []T,
	city string,
	sample SampleFunc[T],
	divideWind bool,
	loc *time.Location,
) models.Forecast {
	var morning, day, evening, night GenericSumData

	localNow := time.Now().In(loc)
	todayStr := localNow.Format("2006-01-02")
	tomorrowStr := localNow.Add(24 * time.Hour).Format("2006-01-02")

	for _, elem := range data {
		dt, temp, feelsLike, humidity, wind, ok := sample(elem)
		if !ok {
			continue
		}

		date := dt.Format("2006-01-02")
		hour := dt.Hour()

		switch {
		case date == todayStr && hour >= 6 && hour < 11:
			//fmt.Println("Добавляем в утро")
			morning.AddSample(temp, feelsLike, humidity, wind)
		case date == todayStr && hour >= 12 && hour <= 15:
			//fmt.Println("Добавляем в день")
			day.AddSample(temp, feelsLike, humidity, wind)
		case date == todayStr && hour >= 18 && hour <= 21:
			//fmt.Println("Добавляем в вечер")
			evening.AddSample(temp, feelsLike, humidity, wind)
		case date == todayStr && hour >= 22:
			//fmt.Println("Добавляем в ночь")
			night.AddSample(temp, feelsLike, humidity, wind)
		case date == tomorrowStr && hour <= 3:
			//fmt.Println("Добавляем в ночь")
			night.AddSample(temp, feelsLike, humidity, wind)
		}
	}

	// возвращаем финальный прогноз с усредненными данными
	return models.Forecast{
		City:    city,
		Morning: morning.Average(divideWind),
		Day:     day.Average(divideWind),
		Evening: evening.Average(divideWind),
		Night:   night.Average(divideWind),
	}
}
