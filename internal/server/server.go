package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/SunMaungOo/task-runner-auth/internal/handler"
	"github.com/SunMaungOo/task-runner-auth/internal/service"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func New(logger *slog.Logger, auth *service.AuthService, authHandler *handler.AuthHandler, readyChecks ...handler.ReadinessCheck) http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handler.Heathz)
	mux.HandleFunc("GET /readyz", handler.Readyz(readyChecks...))
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)

	return recoverMiddleware(logger)(loggingMiddleware(logger)(mux))
}

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(writer http.ResponseWriter, rawRequest *http.Request) {

			start := time.Now()

			statusWriter := &statusWriter{
				ResponseWriter: writer,
				status:         http.StatusOK,
			}

			next.ServeHTTP(statusWriter, rawRequest)

			logger.Info("request",
				"method", rawRequest.Method,
				"path", rawRequest.URL.Path,
				"status", statusWriter.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})

	}
}

// turn a panic in single handler into http status 500 instead of crashing
func recoverMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(writer http.ResponseWriter, rawRequest *http.Request) {

			defer func() {

				if err := recover(); err != nil {

					logger.Error("panic recovered", "error", err, "path", rawRequest.URL.Path)

					http.Error(writer, "internal server error", http.StatusInternalServerError)

				}
			}()

			next.ServeHTTP(writer, rawRequest)

		})

	}
}
