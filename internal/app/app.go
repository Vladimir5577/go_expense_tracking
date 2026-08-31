package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"go_expense_service/internal/config"
	"go_expense_service/internal/handler"
	"go_expense_service/internal/middleware"
	"go_expense_service/internal/repository"
	"go_expense_service/internal/service"
)

type App struct {
	router *chi.Mux
	cfg    *config.Config
}

// Handlers — набор хендлеров, который получает роутер.
type Handlers struct {
	Auth     *handler.AuthHandler
	Category *handler.CategoryHandler
	Expense  *handler.ExpenseHandler
	Report   *handler.ReportHandler
}

// NewApp собирает зависимости вручную: репозитории → сервисы → хендлеры.
// DI-контейнер здесь не нужен, порядок сборки читается сверху вниз.
func NewApp(cfg *config.Config, db *sql.DB) *App {
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)

	authSvc := service.NewAuthService(userRepo, cfg)
	categorySvc := service.NewCategoryService(categoryRepo, cfg)
	expenseSvc := service.NewExpenseService(expenseRepo, categoryRepo, cfg)
	reportSvc := service.NewReportService(expenseRepo, cfg)

	h := Handlers{
		Auth:     handler.NewAuthHandler(authSvc),
		Category: handler.NewCategoryHandler(categorySvc),
		Expense:  handler.NewExpenseHandler(expenseSvc),
		Report:   handler.NewReportHandler(reportSvc),
	}

	authMw := middleware.NewAuthMiddleware(cfg.JWTSecret)

	return &App{
		router: setupRouter(h, authMw, cfg.WebDir),
		cfg:    cfg,
	}
}

func (a *App) Run() error {
	srv := &http.Server{
		Addr:         a.cfg.Host + ":" + a.cfg.Port,
		Handler:      a.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("ошибка HTTP-сервера: %w", err)
	case <-quit:
		slog.Info("Получен сигнал завершения, начинаем graceful shutdown...")
	}

	// Даём 5 секунд на завершение текущих запросов.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("ошибка при остановке сервера: %w", err)
	}

	slog.Info("HTTP-сервер успешно остановлен")
	return nil
}
