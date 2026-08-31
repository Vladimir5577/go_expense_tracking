package app

import (
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"go_expense_service/internal/apperr"
	"go_expense_service/internal/helper"
	"go_expense_service/internal/middleware"
)

func setupRouter(h Handlers, authMw *middleware.AuthMiddleware, webDir string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestLogger())
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(15 * time.Second))

	// Публичная часть
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"expenses"}`))
	})
	r.Post("/api/auth/login", h.Auth.Login())

	// Всё остальное — только с валидным токеном
	r.Group(func(r chi.Router) {
		r.Use(authMw.Handler)

		r.Get("/api/me", h.Auth.Me())

		r.Route("/api/categories", func(r chi.Router) {
			r.Get("/", h.Category.List())
			r.Post("/", h.Category.Create())
			r.Get("/{id}", h.Category.Get())
			r.Patch("/{id}", h.Category.Update())
			r.Delete("/{id}", h.Category.Delete())
		})

		r.Route("/api/expenses", func(r chi.Router) {
			r.Get("/", h.Expense.List())
			r.Post("/", h.Expense.Create())
			r.Get("/{id}", h.Expense.Get())
			r.Patch("/{id}", h.Expense.Update())
			r.Delete("/{id}", h.Expense.Delete())
		})

		r.Get("/api/reports/summary", h.Report.Summary())
	})

	// Статика фронтенда — последним, чтобы не перехватывать маршруты API.
	r.Handle("/*", spaHandler(webDir))

	return r
}

// spaHandler отдаёт файл из webDir, а если файла нет — index.html.
//
// Фолбэк нужен для History-роутинга: прямой заход на /expenses и перезагрузка
// страницы уходят на сервер обычным GET, и без фолбэка вернулся бы 404.
// При хешевом роутинге (#/expenses) фолбэк просто не срабатывает.
func spaHandler(webDir string) http.HandlerFunc {
	root := http.Dir(webDir)
	fileServer := http.FileServer(root)
	index := filepath.Join(webDir, "index.html")

	return func(w http.ResponseWriter, r *http.Request) {
		// Неизвестный путь под /api — это ошибка API, а не маршрут фронта.
		// Без этой проверки клиент вместо JSON-ошибки получил бы HTML-страницу
		// со статусом 200, и причину сбоя пришлось бы искать вслепую.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			helper.WriteError(w, apperr.ErrNotFound)
			return
		}

		// http.Dir сам отсекает выход за пределы каталога — своей проверки пути
		// не пишем, чтобы не разойтись с ней в мелочах.
		f, err := root.Open(path.Clean(r.URL.Path))
		if err != nil {
			http.ServeFile(w, r, index)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	}
}
