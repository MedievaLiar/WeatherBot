# PiterMoment

🪁 Telegram: @weather_aggregat_bot

Бот, собирающий суперпрогноз, используя агрегаторы погоды 😎🌦

# 🌟 Основные возможности

1) Прогноз погоды **высокой точности** за счет использования данных из множества источников:
 - [Gismeteo (HTML-парсинг)](https://www.gismeteo.ru/)
 - [Яндекс.Погода (HTML-парсинг)](https://yandex.ru/pogoda/ru/fiorentino?via=reg&lat=43.909068&lon=12.459716)
 - [OpenWeatherMap API](https://openweathermap.org)
 - [AccuWeather API](https://www.accuweather.com/)
 - [Sunrise-Sunset API](https://sunrise-sunset.org/)
 - [NASA (шутка)]()

2) **Гибкие форматы запросов**:
  - Погода сейчас
  - Прогноз на сегодня
  - Прогноз на завтра

3) **Авторассылка** по расписанию:
  - Настройка времени получения прогноза
  - Автоподстройка под часовой пояс пользователя
4) **Пример прогноза**:
```
📍Погода сегодня, 29.05.2025, в городе:
Санкт-Петербург

🍀 Утро:
Температура: 19°C
(ощущается как 17°C)
Влажность: 73%
Ветер: 5.0 м/с
 ☔️🌧️ Небольшой дождь

🏵 День:
Температура: 23°C
(ощущается как 21°C)
Влажность: 51%
Ветер: 5.0 м/с
 ☔️🌧️ Небольшой дождь

🪁 Вечер:
Температура: 21°C
(ощущается как 19°C)
Влажность: 70%
Ветер: 2.7 м/с
 🌤️ Облачно с прояснениями

🌑 Ночь:
Температура: 17°C
(ощущается как 17°C)
Влажность: 85%
Ветер: 1.7 м/с
 🐱 Пасмурно

🌊 Рассвет: 03:49
🏜️ Закат: 22:02
```

# 🛠️ Что под капотом?
1) Асихронная сборка данных из всех источников

2) Настраиваемое кэширование результата, позволяющее быстро отдавать результат пользователю, при этом не теряя в свежести прогноза

3) Ответ на запросы пользователя асинхронен, используется read-lock
    Вкупе первые два пункта дают нам предполагаемый rps в 10000+ с возможностью легко масштабировать с помощью облачных решений - достаточно настроить apigateway и балансировщик, например, по username

3) Логирование ошибок в поведении для комфортной разработки

4) Структура проекта понятная и легкая для внедрения новых решений. Выдержана чистая архитектура там, где это необходимо:
    ```weather_bot/
    ├── bot/                   # Telegram-бот и логика общения
    │   ├── messages/          # Шаблоны сообщений
    │   ├── storage/           # Кэш прогнозов и пользователей
    │   └── bot.go             # Главная точка входа в бота
    ├── aggregator/            # Сбор и агрегация данных
    │   ├── format/            # Приведение к общему формату
    │   └── weather_sources/   # Источники погоды
    ├── config/                # Ключи, данные городов
    ├── models/                # Модели данных
    ├── main.go                # Точка запуска
    ├── go.mod / go.sum        # Зависимости
    └── README.md              # You are here
    ```

# 🚀 Как развернуть бота у себя?

1) Требуется Go >= 1.20 - https://go.dev/doc/install
2) Клонировать репозиторий:
```
git clone https://github.com/MedievaLiar/WeatherBot.git
cd WeatherBot
```
3) Установить зависимости:
    ```
    go get github.com/PuerkitoBio/goquery github.com/go-telegram-bot-api/telegram-bot-api/v5 github.com/robfig/cron/v3 gopkg.in/yaml.v3
    ```
4) Получить и записать ключи в `./config/`:
   * config/api_keys.yaml
   * config/accu_keys.yaml

    Пример файлов API ключей можно взять из config/example_.yaml

5) Запуск бота:
   ```
   go run main.go
   ```
# ⚠️ Known issues

- В Яндекс.Погоде используется BDUI, а API платное - гарантия на работу этого источника не предоставляется.
- AccuWeather требует API-ключ для каждого города.
- NASA пока не предоставляет реальных данных (но мы держим их в резерве 😉).
