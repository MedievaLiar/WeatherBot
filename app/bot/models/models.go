package models

// данные для одного времени суток
type PeriodWeather struct {
	Temperature   float64
	FeelsLike     float64
	Humidity      int
	WindSpeed     float64
	Precipitation string
}

// струкутра прогноза
type Forecast struct {
	City    string
	Morning PeriodWeather
	Day     PeriodWeather
	Evening PeriodWeather
	Night   PeriodWeather
	Sunrise string
	Sunset  string
}

// данные пользователя в БД
type UserData struct {
	ID                int64  `db:"id"`                  // ID чата (PRIMARY KEY)
	Username          string `db:"username"`            // @username пользователя
	CurrentCity       string `db:"current_city"`        // город для текущих запросов
	ForecastCity      string `db:"forecast_city"`       // город для ежедневной рассылки
	WantDaily         bool   `db:"want_daily"`          // включена ли рассылка
	ForecastMskHour   int    `db:"forecast_msk_hour"`   // час рассылки по мск
	ForecastLocalHour int    `db:"forecast_local_hour"` // час рассылки по локальному времени
}

// временная структура для управления состоянием пользователя
type UserState struct {
	ExpectingInput string   // ожидаемый параметр
	TempUserData   UserData // временный пользователь
}
