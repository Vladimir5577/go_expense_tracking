package main

import (
	"log/slog"
	"os"

	"go_expense_service/internal/app"
	"go_expense_service/internal/config"
	"go_expense_service/internal/logger"
	"go_expense_service/internal/validator"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Некорректная конфигурация", "error", err)
		os.Exit(1)
	}

	if err := cfg.ValidateForServer(); err != nil {
		slog.Error("Некорректная конфигурация", "error", err)
		os.Exit(1)
	}

	logger.Setup(cfg.Env)
	validator.Init()

	db, err := config.ConnectDB(cfg)
	if err != nil {
		slog.Error("Не удалось подключиться к базе данных", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	application := app.NewApp(cfg, db)

	slog.Info("Сервис учёта расходов запускается",
		"addr", cfg.Host+":"+cfg.Port, "db", cfg.DBPath, "web", cfg.WebDir)
	if err := application.Run(); err != nil {
		slog.Error("Ошибка работы HTTP-сервера", "error", err)
		os.Exit(1)
	}
}
