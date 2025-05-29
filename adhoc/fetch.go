package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"weather_bot/config"
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

	var res struct{ Key string }
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
	keys := make(map[string]string)

	for city, info := range config.CityData {
		key, err := getLocationKey(info.Lat, info.Lon, config.Keys.AccuWeather)
		if err != nil {
			fmt.Printf("❌ %s: %v\n", city, err)
			continue
		}
		fmt.Printf("✅ %s: \"%s\"\n", city, key)
		keys[city] = key
	}

	if err := saveYAML("../config/accu_keys.yaml", keys); err != nil {
		log.Fatalf("Ошибка записи YAML: %v", err)
	}
	fmt.Println("🎉 Ключи сохранены в ../config/accu_keys.yaml")
}

func saveYAML(path string, data any) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

