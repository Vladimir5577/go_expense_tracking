package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger логирует HTTP-запросы через slog.
//
// Успешные запросы вне /api/ в лог не попадают. Статику — стили, полтора
// десятка JS-модулей, иконки — браузер забирает при первой загрузке, дальше
// получает 304, и это около 80% строк, в которых нечего разбирать.
//
// Условие написано по статусу, а не только по пути: если отдача файла всё-таки
// сломается, строка в логе будет. Заодно молчит /healthz, который иначе
// засыпает лог опросами мониторинга.
func RequestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			if ww.Status() < http.StatusBadRequest && !strings.HasPrefix(r.URL.Path, "/api/") {
				return
			}

			slog.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
				slog.String("ip", r.RemoteAddr),
			)
		})
	}
}
