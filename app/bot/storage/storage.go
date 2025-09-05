package storage

import (
	"app/bot/logger"
	"app/bot/models"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// инициализирует соединение с БД и создает таблицу, если нужно
func Init() error {
	dbPath := "bot/storage/bd/bot.db"
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("Не удалось создать директорию для БД: %w", err)
	}

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("не удалось открыть БД: %w", err)
	}

	// cоздаем таблицу
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY,
        username TEXT,
        current_city TEXT DEFAULT '',
        forecast_city TEXT DEFAULT '',
        want_daily BOOLEAN DEFAULT FALSE,
        forecast_msk_hour INTEGER DEFAULT 0,
        forecast_local_hour INTEGER DEFAULT 0
    );
    `
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("не удалось создать таблицу: %w", err)
	}

	log.Println("База данных SQLite инициализирована")
	return nil
}

// возвращает данные пользователя по ID чата
func GetUser(chatID int64) (*models.UserData, error) {
	var user models.UserData
	query := `SELECT * FROM users WHERE id = ?`
	row := db.QueryRow(query, chatID)

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.CurrentCity,
		&user.ForecastCity,
		&user.WantDaily,
		&user.ForecastMskHour,
		&user.ForecastLocalHour,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Пользователь не найден, это нормально
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка при чтении пользователя: %w", err)
	}
	return &user, nil
}

// сохраняет или обновляет данные пользователя
func SaveUser(user *models.UserData) error {
	query := `
    INSERT OR REPLACE INTO users
        (id, username, current_city, forecast_city, want_daily, forecast_msk_hour, forecast_local_hour)
    VALUES (?, ?, ?, ?, ?, ?, ?)
    `
	_, err := db.Exec(
		query,
		user.ID,
		user.Username,
		user.CurrentCity,
		user.ForecastCity,
		user.WantDaily,
		user.ForecastMskHour,
		user.ForecastLocalHour,
	)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении пользователя: %w", err)
	}

	if user.WantDaily {
		logger.LogUserAdded(*user)
	} else {
		logger.LogUserUpdated(*user)
	}

	return nil
}

// возвращает всех пользователей, у которых включена рассылка, для крона
func GetAllUsersForForecast() ([]models.UserData, error) {
	var users []models.UserData
	query := `SELECT * FROM users WHERE want_daily = TRUE`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.UserData
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.CurrentCity,
			&user.ForecastCity,
			&user.WantDaily,
			&user.ForecastMskHour,
			&user.ForecastLocalHour,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func GetUserStats() (total int, withDaily int, err error) {
	// общее количество пользователей
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return 0, 0, fmt.Errorf("ошибка получения общего числа пользователей: %w", err)
	}

	// количество пользователей с включенной рассылкой
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE want_daily = TRUE").Scan(&withDaily)
	if err != nil {
		return total, 0, fmt.Errorf("ошибка получения числа пользователей с рассылкой: %w", err)
	}

	return total, withDaily, nil
}

// возвращает текущий город пользователя из БД
func ExtractCurrentCity(chatID int64) (string, bool) {
	user, err := GetUser(chatID)
	if err != nil {
		log.Printf("Ошибка получения пользователя %d: %v", chatID, err)
		return "", false
	}

	if user == nil || user.CurrentCity == "" {
		return "", false
	}

	return user.CurrentCity, true
}

// возвращает город для прогноза из БД
func ExtractForecastCity(chatID int64) (string, bool) {
	user, err := GetUser(chatID)
	if err != nil {
		log.Printf("Ошибка получения пользователя %d: %v", chatID, err)
		return "", false
	}

	if user == nil || user.ForecastCity == "" {
		return "", false
	}

	return user.ForecastCity, true
}

func UpdateCurrentCity(chatID int64, city string, username string) error {
	user, err := GetUser(chatID)
	if err != nil {
		return err
	}

	if user == nil {
		user = &models.UserData{
			ID:          chatID,
			Username:    username,
			CurrentCity: city,
		}
	} else {
		user.CurrentCity = city
		user.Username = username
	}

	if err := SaveUser(user); err != nil {
		return err
	}

	SetCurrentCityInCache(chatID, city)

	return nil
}

func GetAllUsers() ([]models.UserData, error) {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.UserData
	for rows.Next() {
		var user models.UserData
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.CurrentCity,
			&user.ForecastCity,
			&user.WantDaily,
			&user.ForecastMskHour,
			&user.ForecastLocalHour,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// закрываем соединение с базой
func CloseConnection() {
	if db != nil {
		db.Close()
	}
}
