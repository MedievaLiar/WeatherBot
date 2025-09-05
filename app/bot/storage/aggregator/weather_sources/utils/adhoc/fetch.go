package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"app/bot/config"
	"gopkg.in/yaml.v3"
)

func getLocationKey(lat, lon float64, apiKey string) (string, error) {
	u := "http://dataservice.accuweather.com/locations/v1/cities/geoposition/search"
	q := url.Values{
		"apikey": {apiKey},
		"q":      {fmt.Sprintf("%f,%f", lat, lon)},
	}
	resp, err := http.Get(u + "?" + q.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("неудачный ответ от API: %s", resp.Status)
	}

	var res struct {
		Key string
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	if res.Key == "" {
		return "", fmt.Errorf("ключ не найден")
	}
	return res.Key, nil
}

func main() {
	config.LoadAll()

	// загружаем существующие ключи
	existingKeys, err := loadExitingKeys("../../../../../config/accu_keys.yaml")
	if err != nil {
		log.Printf("Не удалось загрузить существующие ключи, создаем новые %v", err)
		existingKeys = make(map[string]string)
	}

	// находим новые города
	newCities := make(map[string]config.City)
	for city, info := range config.CityData {
		if _, exists := existingKeys[city]; !exists {
			newCities[city] = info
			fmt.Printf("🆕 Найден новый город: %s\n", city)
		}
	}

	// если новых городов нет - выходим
	if len(newCities) == 0 {
		fmt.Println("🎉 Новых городов не найдено!")
		return
	}

	// получаем ключи только для новых городов
	for city, info := range newCities {
		key, err := getLocationKey(info.Lat, info.Lon, config.Keys.AccuWeather)
		if err != nil {
			fmt.Printf("❌ %s: %v (пропускаем)\n", city, err)
			continue
		}
		fmt.Printf("✅ %s: \"%s\"\n", city, key)
		existingKeys[city] = key
	}

	// сохраняем обновленный список
	if err := saveYAML("../../../../../config/accu_keys.yaml", existingKeys); err != nil {
		log.Fatalf("💥 Ошибка записи YAML: %v", err)
	}
	fmt.Printf("\n🎉 Файл обновлен! Добавлено %d новых городов. Всего городов в файле: %d\n",
		len(newCities), len(existingKeys))
}

// загружаем существующие ключи из ямл
func loadExitingKeys(path string) (map[string]string, error) {
	data := make(map[string]string)

	// существует ли файл
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return data, nil
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func saveYAML(path string, data any) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}
