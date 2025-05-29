package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"weather_bot/bot/messages"
	"weather_bot/bot/storage/now"
	"weather_bot/bot/storage/today"
	"weather_bot/bot/storage/tomorrow"
	"weather_bot/bot/storage/users"
	"weather_bot/config"

	"github.com/robfig/cron/v3"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI

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
	var err error
	bot, err = tgbotapi.NewBotAPI(config.Keys.TelegramBotToken)
	if err != nil {
		return err
	}

	fmt.Printf("Бот запущен как @%s\n", bot.Self.UserName)

	users.LoadUserCache()
	go startScheduler()
	go users.StartAutoSave()

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

    if users.IsAwaitingForecastHour(chatID) {
        if localHour, err := strconv.Atoi(msg.Text); err == nil && localHour >= 0 && localHour <= 23 {
			city, _ := users.GetForecastCity(chatID)
			mskHour := localHour
			if config.CityData[city].Timezone != "Europe/Moscow" {
				mskHour = users.ConvertToMsk(city, localHour)
			}

            users.SetUserForecast(chatID, true, mskHour, localHour)
            users.SetAwaitingForecastHour(chatID, false)

			msg := fmt.Sprintf(messages.ForecastConfirmed, city, localHour)
            send(chatID, msg, mainKeyboard)

			return
        } else {
            send(chatID, messages.InvalidTime, nil)
            return
        }
    }

    switch msg.Text {
    case "/start":
        send(chatID, messages.Welcome, mainKeyboard)
    case "🏙️ Выбрать город":
        selectCity(chatID)
    case "📅 Ежедневный прогноз":
        showForecastMenu(chatID)
    case "🌤️ Сейчас":
        getWeather(chatID, now.GetWeatherNow)
    case "☀️ Сегодня":
        getWeather(chatID, today.GetTodayForecast)
    case "🌙 Завтра":
        getWeather(chatID, tomorrow.GetTomorrowForecast)
    default:
        send(chatID, messages.UnknownCommand, mainKeyboard)
    }
}

func showForecastMenu(chatID int64) {
	_, hasForecastCity := users.GetForecastCity(chatID)

    if !hasForecastCity {
        keyboard := tgbotapi.NewInlineKeyboardMarkup(
            tgbotapi.NewInlineKeyboardRow(
                tgbotapi.NewInlineKeyboardButtonData("Настроить ✅", "setup_forecast"),
                tgbotapi.NewInlineKeyboardButtonData("Не сейчас ❌", "decline_forecast"),
            ),
        )
        send(chatID, messages.DailyForecastInfo, keyboard)
        return
    }

    city, _ := users.GetForecastCity(chatID)
    _, hour, _ := users.GetUserForecastPrefs(chatID)

    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("🏝 Сменить город", "change_city"),
            tgbotapi.NewInlineKeyboardButtonData("⏰ Изменить время", "change_time"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("🔌 Отключить", "disable_forecast"),
        ),
    )
    send(chatID, fmt.Sprintf(messages.ForecastSettings, city, hour), keyboard)
}

func handleCallback(callback *tgbotapi.CallbackQuery) {
    chatID := callback.Message.Chat.ID
    data := callback.Data

    bot.Send(tgbotapi.NewCallback(callback.ID, ""))

    switch {
    case strings.HasPrefix(data, "city_"):
        city := strings.TrimPrefix(data, "city_")
		if users.IsChangingCity(chatID) {

			users.SetForecastCity(chatID, city)
			_, hour, _ := users.GetUserForecastPrefs(chatID)
			send(chatID, fmt.Sprintf(messages.ForecastConfirmed, city, hour), mainKeyboard)
			users.WantChangeDailyCity(chatID, false)
			break
		}

        if users.IsAwaitingForecastHour(chatID) {

			users.SetForecastCity(chatID, city)
			send(chatID, fmt.Sprintf(messages.DailyCitySelected, city), nil)
			send(chatID, messages.EnterTime, nil)
        } else {
            users.SetCurrentCity(chatID, city)
            send(chatID, fmt.Sprintf(messages.CitySelected, city), mainKeyboard)
        }

    case data == "decline_forecast":
        send(chatID, messages.SetupDeclined, mainKeyboard)

	case data == "setup_forecast":
    	users.SetAwaitingForecastHour(chatID, true)
    	send(chatID, messages.ForecastSetup, nil)
    	selectCity(chatID)

	case data == "change_city":
		users.WantChangeDailyCity(chatID, true)
    	selectCity(chatID)

	case data == "change_time":
    	users.SetAwaitingForecastHour(chatID, true)
    	send(chatID, messages.EnterTime, nil)

    case data == "disable_forecast":
        msg := DisableForecast(chatID)
        send(chatID, msg, mainKeyboard)

    }
}

func getWeather(chatID int64, weatherFunc func(string) (string, error)) {
    city, ok := users.GetCurrentCity(chatID)
    if !ok {
        selectCity(chatID)
        return
    }

    send(chatID, messages.WeatherFetching(), nil)
    time.Sleep(1 * time.Second)

    if forecast, err := weatherFunc(city); err == nil {
        send(chatID, forecast, mainKeyboard)
    } else {
        send(chatID, messages.WeatherError, mainKeyboard)
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

func startScheduler() {
    c := cron.New()
    c.AddFunc("0 * * * *", sendDailyForecasts)
    c.Start()
}

func sendDailyForecasts() {
    hour := time.Now().Hour()
    for chatID, prefs := range users.GetAllUserPrefs() {
        if prefs.WantDailyForecast && prefs.ForecastMskHour == hour {
            go sendForecast(chatID, prefs.ForecastCity)
        }
    }
}

func sendForecast(chatID int64, city string) {
    send(chatID, messages.YourDailyForecast, nil)

    if forecast, err := today.GetTodayForecast(city); err == nil {
        send(chatID, forecast, nil)
    } else {
        send(chatID, messages.WeatherError, nil)
    }
}

func send(chatID int64, text string, keyboard any) {
    msg := tgbotapi.NewMessage(chatID, text)
    if keyboard != nil {
        msg.ReplyMarkup = keyboard
    }
    bot.Send(msg)
}

func SetupForecast(chatID int64) string {
    users.SetAwaitingForecastHour(chatID, true)
    return messages.ForecastSetup
}

func DisableForecast(chatID int64) string {
    users.SetUserForecast(chatID, false, 0, 0)
    return messages.ForecastDisabled
}

