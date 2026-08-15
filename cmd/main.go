package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/SunMaungOo/task-runner-auth/internal/config"
	"github.com/SunMaungOo/task-runner-auth/internal/handler"
	"github.com/SunMaungOo/task-runner-auth/internal/mongodb"
	"github.com/SunMaungOo/task-runner-auth/internal/server"
	"github.com/SunMaungOo/task-runner-auth/internal/service"
)

func newLogger(level string) *slog.Logger {

	lvl := slog.LevelInfo

	switch strings.ToLower(level) {
	case "debug":

		lvl = slog.LevelDebug

	case "warn", "warning":
		lvl = slog.LevelWarn

	case "error":
		lvl = slog.LevelError

	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

func main() {

	cfg := config.Load()

	logger := newLogger(cfg.LogLevel)

	if cfg.JWTSecret == "" {

		logger.Error("JWT_SECRET is not set - unable to start")

		os.Exit(1)
	}

	connectContext, cancel := context.WithTimeout(context.Background(), cfg.DatabaseConnectTimeout)

	defer cancel()

	userRepo, err := mongodb.New(connectContext, cfg.MongoURI, cfg.MongoDatabase)

	if err != nil {

		logger.Error("failed to connect to MongoDB", "error", err)

		os.Exit(1)
	}

	auth := service.New(userRepo, []byte(cfg.JWTSecret), cfg.TokenTTL, logger)

	authHandler := handler.NewAuthHandler(auth)

	mongoPing := func(request *http.Request) error {

		ctx, cancel := context.WithTimeout(request.Context(), cfg.DatabasePingTimeout)

		defer cancel()

		return userRepo.Ping(ctx)
	}

	handlerChain := server.New(logger, auth, authHandler, mongoPing)

	srv := http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handlerChain,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errorChannel := make(chan error, 1)

	go func() {

		logger.Info("starting server", "port", cfg.Port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {

			errorChannel <- err
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {

	case err := <-errorChannel:

		logger.Error("server failed to start", "error", err)

		os.Exit(1)

	case sig := <-quit:
		logger.Info("shutdown signal received", "signal", sig.String())
	}

	shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)

	defer shutdownCancel()

	if err := srv.Shutdown(shutdownContext); err != nil {
		logger.Error("graceful shutdown failed", "error", err)

		os.Exit(1)
	}

	logger.Info("server stopped cleanly")

}
