package weather_sources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"weather_bot/config"
)

func SaveRawAccuWeatherJSON(city string) error {
	apiKey := config.Keys.AccuWeather
	locationKey := config.AccuLocationKeys[city]
	url := fmt.Sprintf(
		"http://dataservice.accuweather.com/forecasts/v1/hourly/24hour/%s?apikey=%s&language=ru&details=true&metric=true",
		locationKey, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("получен статус %d от AccuWeather", resp.StatusCode)
	}

	var raw any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return fmt.Errorf("ошибка при декодировании JSON: %w", err)
	}

	rawJSON, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге JSON: %w", err)
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("не удалось определить путь к текущему файлу")
	}

	saveDir := filepath.Join(filepath.Dir(thisFile), "tests/accuweather_check/raw_json")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию: %w", err)
	}

	filename := fmt.Sprintf("accuweather_raw_%s.json", time.Now().Format("2006-01-02_15-04"))
	fullPath := filepath.Join(saveDir, filename)

	if err := os.WriteFile(fullPath, rawJSON, 0644); err != nil {
		return fmt.Errorf("ошибка при сохранении файла: %w", err)
	}

	fmt.Println("JSON успешно сохранен:", fullPath)
	return nil
}

