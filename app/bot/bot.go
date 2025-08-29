package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"app/bot/config"
	"app/bot/logger"
	"app/bot/messages"
	"app/bot/models"
	"app/bot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
)

var (
	bot        *tgbotapi.BotAPI
	userStates = make(map[int64]*models.UserState) // собираем состояние пользователя
)

var mainKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🌤️ Сейчас"),
		tgbotapi.NewKeyboardButton("☀️ Сегодня"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("🌙 Завтра"),
		tgbotapi.NewKeyboardButton("🏙️ Выбрать город"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("📅 Ежедневный прогноз"),
	),
)

func Start() error {
	// запуск логирования
	if err := logger.Init(); err != nil {
		log.Printf("Ошибка инициализации логирования: %v", err)
	}
	defer logger.Close()

	logger.StartRotation()

	var err error
	bot, err = tgbotapi.NewBotAPI(config.Keys.TelegramBotToken)
	if err != nil {
		return err
	}

	fmt.Printf("Бот запущен как @%s\n", bot.Self.UserName)

	go startScheduler()

	// подгрузка БД пользователей
	if err := storage.Init(); err != nil {
		return fmt.Errorf("ошибка инициализации БД: %v", err)
	}
	defer storage.CloseConnection()

	totalUsers, usersWithDaily, err := storage.GetUserStats()
	if err != nil {
		log.Printf("Ошибка получения статистики пользователей: %v", err)
	} else {
		fmt.Printf("[Database] Загружено %d пользователей\n", totalUsers)
		fmt.Printf("[Database] Из них с ежедневной рассылкой: %d\n", usersWithDaily)
	}

	updates := bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
	for update := range updates {
		switch {
		case update.Message != nil:
			handleMessage(update.Message)
		case update.CallbackQuery != nil:
			handleCallback(update.CallbackQuery)
		}
	}

	return nil
}

func handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := msg.Text
	username := msg.Chat.UserName

	if state, exists := userStates[chatID]; exists {
		handleUserInput(chatID, state, text, username)
		return
	}

	switch text {
	case "/start":
		user, _ := storage.GetUser(chatID)
		if user == nil {
			send(chatID, messages.Welcome, mainKeyboard)
		}
	case "🏙️ Выбрать город":
		askForCity(chatID, "current_city")
	case "📅 Ежедневный прогноз":
		showForecastMenu(chatID)
	case "🌤️ Сейчас":
		getWeather(chatID, storage.GetNowForecast)
	case "☀️ Сегодня":
		getWeather(chatID, storage.GetTodayForecast)
	case "🌙 Завтра":
		getWeather(chatID, storage.GetTomorrowForecast)
	default:
		send(chatID, messages.UnknownCommand, mainKeyboard)
	}
}

func handleUserInput(chatID int64, state *models.UserState, text string, username string) {
	switch state.ExpectingInput {
	case "forecast_local_hour":
		// ждем установку времени
		localHour, err := strconv.Atoi(text)
		if err != nil || localHour < 0 || localHour > 23 {
			send(chatID, messages.InvalidTime, nil)
			return
		}
		state.TempUserData.ForecastLocalHour = localHour
		mskHour := convertToMsk(state.TempUserData.ForecastCity, localHour)
		state.TempUserData.ForecastMskHour = mskHour
		state.TempUserData.WantDaily = true

		state.TempUserData.ID = chatID
		state.TempUserData.Username = username

		// сохраняем в БД
		err = storage.SaveUser(&state.TempUserData)
		if err != nil {
			log.Printf("Ошибка сохранения пользователя %d: %v", chatID, err)
			send(chatID, "Произошла ошибка при настройке", nil)
		} else {
			msg := fmt.Sprintf(messages.ForecastConfirmed, state.TempUserData.ForecastCity, localHour)
			send(chatID, msg, mainKeyboard)
		}
		// удаляем временного пользователя
		delete(userStates, chatID)

	default:
		send(chatID, "Пожалуйста, используйте кнопки для выбора", mainKeyboard)
	}
}

func handleCallback(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	bot.Send(tgbotapi.NewCallback(callback.ID, ""))

	if city, found := strings.CutPrefix(data, "city_"); found {
		if state, exists := userStates[chatID]; exists {
			switch state.ExpectingInput {
			case "current_city":
				err := storage.UpdateCurrentCity(chatID, city, callback.Message.Chat.UserName)
				if err != nil {
					log.Printf("Ошибка сохранения города: %v", err)
					send(chatID, "Ошибка сохранения города", mainKeyboard)
				} else {
					send(chatID, fmt.Sprintf(messages.CitySelected, city), mainKeyboard)
				}
				delete(userStates, chatID)

			case "forecast_city":
				state.TempUserData.ForecastCity = city
				state.ExpectingInput = "forecast_local_hour" // меняем состояние на ожидание часа
				send(chatID, fmt.Sprintf(messages.DailyCitySelected, city), nil)
				send(chatID, messages.EnterTime, nil)
			}
			return
		}
		err := storage.UpdateCurrentCity(chatID, city, callback.Message.Chat.UserName)
		if err != nil {
			log.Printf("Ошибка сохранения города: %v", err)
			send(chatID, "Ошибка сохранения города", mainKeyboard)
		} else {
			send(chatID, fmt.Sprintf(messages.CitySelected, city), mainKeyboard)
		}
		return
	}

	switch data {
	case "setup_forecast":
		// начинаем процесс настройки рассылки
		state := &models.UserState{ExpectingInput: "forecast_city"}
		userStates[chatID] = state
		askForCity(chatID, "forecast_city")

	case "change_city":
		// меняем город для рассылки
		state := &models.UserState{ExpectingInput: "forecast_city"}
		userStates[chatID] = state
		askForCity(chatID, "forecast_city")

	case "change_time":
		// меняем время для рассылки
		state, exists := userStates[chatID]
		if !exists {
			state = &models.UserState{}
			userStates[chatID] = state
		}
		// загружаем текущие настройки пользователя
		user, _ := storage.GetUser(chatID)
		if user != nil {
			state.TempUserData = *user
		}
		state.ExpectingInput = "forecast_local_hour"
		send(chatID, messages.EnterTime, nil)

	case "disable_forecast":
		user, err := storage.GetUser(chatID)
		if err != nil {
			log.Printf("Ошибка получения пользователя %d: %v", chatID, err)
			send(chatID, "Ошибка отключения рассылки", mainKeyboard)
			return
		}
		if user != nil {
			user.WantDaily = false
			if err := storage.SaveUser(user); err != nil {
				log.Printf("Ошибка сохранения пользователя %d: %v", chatID, err)
				send(chatID, "Ошибка отключения рассылки", mainKeyboard)
			} else {
				send(chatID, messages.ForecastDisabled, mainKeyboard)
			}
		} else {
			send(chatID, "Настройки рассылки не найдены", mainKeyboard)
		}
	}
}

func getWeather(chatID int64, weatherFunc func(string) (string, error)) {
	city, ok := storage.GetCurrentCity(chatID)
	if !ok {
		selectCity(chatID)
		return
	}

	send(chatID, messages.WeatherFetching(), nil)

	if forecast, err := weatherFunc(city); err == nil {
		send(chatID, forecast, mainKeyboard)
	} else {
		send(chatID, messages.WeatherError, mainKeyboard)
	}
}

// выводим меню с настройками рассылки
func showForecastMenu(chatID int64) {
	user, err := storage.GetUser(chatID)
	if err != nil {
		send(chatID, "Ошибка получения данных", mainKeyboard)
		return
	}

	hasActiveForecast := user != nil && user.WantDaily && user.ForecastCity != ""

	if !hasActiveForecast {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Настроить ✅", "setup_forecast"),
				tgbotapi.NewInlineKeyboardButtonData("Не сейчас ❌", "decline_forecast"),
			),
		)
		send(chatID, messages.DailyForecastInfo, keyboard)
		return
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏝 Сменить город", "change_city"),
			tgbotapi.NewInlineKeyboardButtonData("⏰ Изменить время", "change_time"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔌 Отключить", "disable_forecast"),
		),
	)
	send(chatID, fmt.Sprintf(messages.ForecastSettings, user.ForecastCity, user.ForecastLocalHour), keyboard)
}

func askForCity(chatID int64, purpose string) {
	selectCity(chatID)

	state, exists := userStates[chatID]
	if !exists {
		state = &models.UserState{}
		userStates[chatID] = state
	}
	state.ExpectingInput = purpose

	if purpose == "forecast_city" {
		if user, err := storage.GetUser(chatID); err == nil && user != nil {
			state.TempUserData = *user
		}
	}
}

func send(chatID int64, text string, keyboard any) {
	msg := tgbotapi.NewMessage(chatID, text)
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}
	bot.Send(msg)
}

func startScheduler() {
	c := cron.New()
	c.AddFunc("0 * * * *", sendDailyForecasts) // каждый час
	c.Start()
}

func sendDailyForecasts() {
	// берем текущий час по мск
	currentHour := time.Now().UTC().Add(3 * time.Hour).Hour() // UTC+3 = MSK
	// получаем всех пользователей с рассылкой из БД
	users, err := storage.GetAllUsersForForecast()
	if err != nil {
		log.Printf("Ошибка получения пользователей для рассылки: %v", err)
		return
	}

	for _, user := range users {
		// рассылаем в нужный час
		if user.ForecastMskHour == currentHour {
			go sendForecast(user.ID, user.ForecastCity)
		}
	}
}

func selectCity(chatID int64) {
	var buttons [][]tgbotapi.InlineKeyboardButton
	for city := range config.CityData {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(city, "city_"+city),
		))
	}

	msg := tgbotapi.NewMessage(chatID, messages.SelectCity)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons...)
	bot.Send(msg)
}

func sendForecast(chatID int64, city string) {
	send(chatID, messages.YourDailyForecast, nil)

	if forecast, err := storage.GetTodayForecast(city); err == nil {
		send(chatID, forecast, nil)
	} else {
		send(chatID, messages.WeatherError, nil)
	}
}

func convertToMsk(city string, localHour int) int {
	loc, _ := time.LoadLocation(config.CityData[city].Timezone)

	localTime := time.Date(2006, 1, 2, localHour, 0, 0, 0, loc)

	mskLoc, _ := time.LoadLocation("Europe/Moscow")
	mskTime := localTime.In(mskLoc)
	return mskTime.Hour()
}
