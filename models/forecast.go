package models

type PeriodWeather struct {
	Temperature   float64
	FeelsLike     float64
	Humidity      int
	WindSpeed     float64
	Precipitation string
}

type Forecast struct {
	City    string
	Morning PeriodWeather
	Day 	PeriodWeather
	Evening PeriodWeather
	Night   PeriodWeather
	Sunrise string
	Sunset  string
}
