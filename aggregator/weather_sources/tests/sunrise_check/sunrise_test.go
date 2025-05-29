package main

import (
	"testing"
	"time"

	"weather_bot/aggregator/weather_sources"
)

func TestGetSunriseSunset(t *testing.T) {
	city := "Moscow"
	data, err := weather_sources.GetSunriseSunset(city, true)
	if err != nil {
		t.Fatalf("ошибка при получении времени заката/рассвета: %v", err)
	}

	layout := "15:04"

	t.Logf("Raw sunrise: %s, sunset: %s", data.Sunrise, data.Sunset)

	sunrise, err := time.Parse(layout, data.Sunrise)
	if err != nil {
		t.Fatalf("не удалось распарсить рассвет: %v", err)
	}

	sunset, err := time.Parse(layout, data.Sunset)
	if err != nil {
		t.Fatalf("не удалось распарсить закат: %v", err)
	}

	if sunrise.IsZero() || sunset.IsZero() {
		t.Error("рассвет или закат равен нулю")
	}

	if !sunrise.Before(sunset) {
		t.Errorf("рассвет (%v) должен быть раньше заката (%v)", sunrise, sunset)
	}

	if sunrise.Hour() < 2 || sunrise.Hour() > 10 {
		t.Errorf("подозрительное время рассвета: %v", sunrise)
	}

	//проверка что закат в правильное время
	if sunset.Hour() < 15 || sunset.Hour() > 23 {
		t.Errorf("подозрительное время заката: %v", sunset)
	}

	t.Logf("Рассвет: %v, Закат: %v", sunrise.Format("15:04"), sunset.Format("15:04"))
}
