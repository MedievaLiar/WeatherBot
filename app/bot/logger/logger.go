package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"app/bot/models"
)

// храним состояние логгера
var (
	logFile     *os.File // текущий открытый файл лога
	logger      *log.Logger
	currentDate string
	mutex       sync.Mutex
)

// Инициализирует систему логирования
func Init() error {
	return createOrOpenLogFile()
}

func createOrOpenLogFile() error {
	mutex.Lock()
	defer mutex.Unlock()

	logsDir := "bot/logger/logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию логов: %w", err)
	}

	today := time.Now().Format("2006-01-02")

	// если текущий файл подходит для даты, то выходим
	if logFile != nil && currentDate == today {
		return nil
	}

	// закрываем предыдущий файл лога (смена даты)
	if logFile != nil {
		logFile.Close()
	}

	// создаем новый файл лога
	logPath := filepath.Join(logsDir, fmt.Sprintf("bot_%s.log", today))
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл лога: %w", err)
	}

	// добавляем текущую дату и файл+строку
	logger = log.New(file, "", log.LstdFlags|log.Lshortfile)
	logFile = file
	currentDate = today

	logger.Printf("Система логирования инициализирована для даты %s", today)
	return nil
}

func checkDate() {
	today := time.Now().Format("2006-01-02")
	if currentDate != today {
		createOrOpenLogFile()
	}
}

// добавление пользователя
func LogUserAdded(user models.UserData) {
	checkDate()

	message := fmt.Sprintf("ДОБАВЛЕН ПОЛЬЗОВАТЕЛЬ: ID=%d, Username=@%s, CurrentCity=%s, ForecastCity=%s, WantDaily=%t",
		user.ID, user.Username, user.CurrentCity, user.ForecastCity, user.WantDaily)

	logger.Println(message)
	fmt.Println(message) // Также выводим в консоль
}

// обновление пользователя
func LogUserUpdated(user models.UserData) {
	checkDate()

	message := fmt.Sprintf("ОБНОВЛЕН ПОЛЬЗОВАТЕЛЬ: ID=%d, Username=@%s, CurrentCity=%s, ForecastCity=%s, WantDaily=%t",
		user.ID, user.Username, user.CurrentCity, user.ForecastCity, user.WantDaily)

	logger.Println(message)
	fmt.Println(message)
}

// универсальное логирование ошибок
func Error(format string, v ...any) {
	checkDate()
	message := fmt.Sprintf("ERROR: "+format, v...)
	logger.Println(message)
	fmt.Println(message)
}

func GetAvailableLogs() ([]string, error) {
	logsDir := "bot/logger/logs"
	files, err := os.ReadDir(logsDir)
	if err != nil {
		return nil, err
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".log" {
			logFiles = append(logFiles, file.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(logFiles)))

	if len(logFiles) > 3 {
		logFiles = logFiles[:3]
	}

	return logFiles, nil
}

func GetLogFile(filename string) ([]byte, error) {
	logsDir := "bot/logger/logs"
	path := filepath.Join(logsDir, filename)
	return os.ReadFile(path)
}

// закрываем файл лога
func Close() {
	mutex.Lock()
	defer mutex.Unlock()

	if logFile != nil {
		logFile.Close()
		logFile = nil
		currentDate = ""
	}
}
