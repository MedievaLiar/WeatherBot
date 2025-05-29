package format

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"weather_bot/config"
	"weather_bot/models"
)

func FormatNowWeather(city string, p models.PeriodWeather) string {
	temp := int(math.Round(p.Temperature))
	feels := int(math.Round(p.FeelsLike))

	output := fmt.Sprintf(
		"🪂 Погода сейчас в городе:\n"+
			"%s\n\n"+
			"🍀 Температура: %d°C\n"+
			"🎭 (ощущается как %d°C)\n"+
			"💧 Влажность: %d%%\n"+
			"🪁 Ветер: %.1f м/с",
		city, temp, feels, p.Humidity, p.WindSpeed,
	)

	if p.Precipitation != "" {
		output += fmt.Sprintf("\n\n☔️ %s", p.Precipitation)
	}

	output += fmt.Sprintf("\n\nВремя: %s", getLocalTime(city))

	return output
}

func FormatTodayWeather(f models.Forecast) string {
	return fmt.Sprintf(
		"📍Погода сегодня, %s, в городе:\n"+
			"%s\n\n"+
			"🍀 Утро:\n%s\n\n"+
			"🏵 День:\n%s\n\n"+
			"🪁 Вечер:\n%s\n\n"+
			"🌑 Ночь:\n%s\n\n"+
			"🌊 Рассвет: %s\n"+
			"🏜️ Закат: %s",
		getTodayDate(f.City), f.City,
		formatPeriod(f.Morning),
		formatPeriod(f.Day),
		formatPeriod(f.Evening),
		formatPeriod(f.Night),
		f.Sunrise, f.Sunset,
	)
}

func FormatTomorrowWeather(f models.Forecast) string {
	return fmt.Sprintf(
		"📍Погода завтра, %s, в городе:\n"+
			"%s\n\n"+
			"🍀 Утро:\n%s\n\n"+
			"🏵 День:\n%s\n\n"+
			"🪁 Вечер:\n%s\n\n"+
			"🌑 Ночь:\n%s\n\n"+
			"🌊 Рассвет: %s\n"+
			"🏜️ Закат: %s",
			/*"🌄 Утро:\n%s\n\n"+
			"☀️ День:\n%s\n\n"+
			"🎆 Вечер:\n%s\n\n"+
			"🌑 Ночь:\n%s\n\n"+
			"🌅 Рассвет: %s\n"+
			"🌇 Закат: %s",*/
		getTomorrowDate(f.City), f.City,
		formatPeriod(f.Morning),
		formatPeriod(f.Day),
		formatPeriod(f.Evening),
		formatPeriod(f.Night),
		f.Sunrise, f.Sunset,
	)
}

func formatPeriod(p models.PeriodWeather) string {
	temp := int(math.Round(p.Temperature))
	feels := int(math.Round(p.FeelsLike))

	output := fmt.Sprintf(
		"Температура: %d°C\n"+
			"(ощущается как %d°C)\n"+
			"Влажность: %d%%\n"+
			"Ветер: %.1f м/с",
		temp, feels, p.Humidity, p.WindSpeed,
	)

	if p.Precipitation != "" {
		output += fmt.Sprintf("\n %s", precipEmoji(p.Precipitation))
	}

	return output
}

func getTodayDate(city string) string {
	loc, _ := time.LoadLocation(config.CityData[city].Timezone)
	return time.Now().In(loc).Format("02.01.2006")
}

func getTomorrowDate(city string) string {
	loc, _ := time.LoadLocation(config.CityData[city].Timezone)
	return time.Now().In(loc).Add(24 * time.Hour).Format("02.01.2006")
}

func getLocalTime(city string) string {
	loc, _ := time.LoadLocation(config.CityData[city].Timezone)
	return time.Now().In(loc).Format("15:04:05")
}

func precipEmoji(desc string) string {
	switch {
	case strings.Contains(desc, "ясно") || strings.Contains(desc, "малооблачно"):
		return "😎 " + toUp(desc)
	case strings.Contains(desc, "облачно с прояснениями"):
		return "🌤️ " + toUp(desc)
	case strings.Contains(desc, "облачно"):
		return getRandomDullyEmoji() + " " + toUp(desc)
	case strings.Contains(desc, "пасмурно"):
		return getRandomDullyEmoji() + " " + toUp(desc)
	case strings.Contains(desc, "дождь с грозой") || strings.Contains(desc, "гроза"):
		return "⛈️⚡ " + toUp(desc)
	case strings.Contains(desc, "дождь") || strings.Contains(desc, "ливень") || strings.Contains(desc, "небольшой дождь"):
		return "☔️🌧️ " + toUp(desc)
	case strings.Contains(desc, "снег"):
		return "❄️ " + toUp(desc)
	case strings.Contains(desc, "метель"):
		return "🌨️🌪️ " + toUp(desc)
	case strings.Contains(desc, "туман"):
		return "🌫️👀️🐾 " + toUp(desc)
	case strings.Contains(desc, "град"):
		return "🧊 " + toUp(desc)
	default:
		return "🤷‍♀️❓ " + toUp(desc)
	}
}

func getRandomDullyEmoji() string {
	emojis := []string{
		"🧸", "☕", "🐸", "🌿", "🐱", "💤", "🥐", "📖",
		"🪟", "📺", "🍿","🎧", "🧩", "🐌", "🫧", "🎮", "🧃",
	}
	return emojis[rand.Intn(len(emojis))]
}

func toUp(s string) string {
	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}
