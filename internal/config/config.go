package config

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	// Вшивает базу таймзон в бинарник (~450 КБ).
	//
	// Без неё time.LoadLocation зависит от /usr/share/zoneinfo, которого нет
	// ни в alpine, ни в scratch, ни в slim-образах — и сервис падает на старте
	// с "unknown time zone Europe/Moscow". Импорт стоит здесь, а не в main:
	// таймзона нужна тому, кто её загружает, и так её получают оба бинарника.
	_ "time/tzdata"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"

	"go_expense_service/internal/helper"
)

type Config struct {
	Env string

	// Host — интерфейс, на котором слушает сервис. Пусто = все интерфейсы.
	// За обратным прокси нужно ставить 127.0.0.1, иначе сервис доступен
	// снаружи напрямую, в обход TLS.
	Host string
	Port string

	DBPath string

	// WebDir — папка со статикой фронтенда. Путь берём из конфига, а не
	// хардкодим: относительный путь считается от рабочей директории, и запуск
	// бинарника из другого места иначе ломает отдачу файлов.
	WebDir string

	JWTSecret string
	JWTTTL    time.Duration

	// Timezone решает, какой день считается «сегодня» и как считаются границы
	// периодов (день/неделя/месяц).
	Timezone *time.Location

	BcryptCost        int
	LoginMaxAttempts  int
	LoginLockDuration time.Duration

	Clock helper.Clock
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Файл .env не найден, используются системные переменные окружения")
	}

	tzName := getEnv("APP_TIMEZONE", "Europe/Moscow")
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("некорректная APP_TIMEZONE %q: %w", tzName, err)
	}

	c := &Config{
		Env:  getEnv("ENV", "local"),
		Host: getEnv("SERVER_HOST", ""),
		Port: getEnv("SERVER_PORT", "8080"),

		DBPath: getEnv("DB_PATH", "./data/expenses.db"),
		WebDir: getEnv("WEB_DIR", "./web"),

		JWTSecret: getEnv("JWT_SECRET", ""),
		JWTTTL:    time.Duration(getEnvAsInt("JWT_TTL_HOURS", 168)) * time.Hour,

		Timezone: loc,

		BcryptCost:        getEnvAsInt("BCRYPT_COST", 12),
		LoginMaxAttempts:  getEnvAsInt("LOGIN_MAX_ATTEMPTS", 5),
		LoginLockDuration: time.Duration(getEnvAsInt("LOGIN_LOCK_MINUTES", 15)) * time.Minute,

		Clock: helper.NewClock(),
	}

	return c, nil
}

// ValidateForServer проверяет то, что нужно только HTTP-сервису.
// Консольной утилите useradm секрет подписи не нужен, поэтому проверка вынесена
// из Load — иначе useradm нельзя было бы запустить без JWT_SECRET.
func (c *Config) ValidateForServer() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET обязателен")
	}
	return nil
}

// ConnectDB открывает SQLite и настраивает пул.
//
// PRAGMA задаются в DSN, а не разовым Exec: они действуют на соединение, а пул
// открывает новые соединения по мере надобности. Особенно это важно для
// foreign_keys — иначе ON DELETE RESTRICT молча не сработает на части запросов.
//
// _txlock=immediate: по умолчанию BEGIN отложенный, и транзакция вида
// «сначала прочитал, потом пишу» может получить SQLITE_BUSY_SNAPSHOT, который
// busy_timeout не лечит. С immediate блокировка берётся сразу на BEGIN и
// busy_timeout работает как задумано.
func ConnectDB(cfg *Config) (*sql.DB, error) {
	dsn := "file:" + cfg.DBPath +
		"?_pragma=journal_mode(WAL)" +
		"&_pragma=busy_timeout(5000)" +
		"&_pragma=foreign_keys(1)" +
		"&_pragma=synchronous(NORMAL)" +
		"&_txlock=immediate"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("открытие БД: %w", err)
	}

	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(0)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("проверка соединения с БД: %w", err)
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if v, err := strconv.Atoi(getEnv(key, "")); err == nil {
		return v
	}
	return fallback
}
