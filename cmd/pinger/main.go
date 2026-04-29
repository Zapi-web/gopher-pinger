package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	handDelete "github.com/Zapi-web/gopher-pinger/internal/api/handlers/delete"
	"github.com/Zapi-web/gopher-pinger/internal/api/handlers/get"
	logMiddleware "github.com/Zapi-web/gopher-pinger/internal/api/handlers/middleware"
	"github.com/Zapi-web/gopher-pinger/internal/api/handlers/post"
	"github.com/Zapi-web/gopher-pinger/internal/api/handlers/put"
	"github.com/Zapi-web/gopher-pinger/internal/config"
	"github.com/Zapi-web/gopher-pinger/internal/logger"
	"github.com/Zapi-web/gopher-pinger/internal/metrics"
	"github.com/Zapi-web/gopher-pinger/internal/service"
	"github.com/Zapi-web/gopher-pinger/internal/storage/database"
	"github.com/Zapi-web/gopher-pinger/internal/storage/local"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	cfg, err := config.Init()

	if err != nil {
		slog.Error("failed to read config", "err", err)
		os.Exit(1)
	}

	slog.SetDefault(logger.New(cfg.LogLevel))
	slog.Info("logger initialized", "level", cfg.LogLevel)

	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	state, err := database.New(cfg.Addr)

	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer state.Close()

	processes := local.InitMap()
	promMetrics := metrics.New()

	metricsInterface := service.NewMetricsService(*promMetrics)
	controlInterface := service.NewService(appCtx, processes, state, metricsInterface)

	err = controlInterface.Init()
	go controlInterface.ResultsMonitoring()

	if err != nil {
		slog.Error("failed to init service", "err", err)
		return
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return logMiddleware.LoggingMiddleware(next, metricsInterface)
	})

	r.Post("/newPinger", post.New(controlInterface))
	r.Get("/getPinger", get.New(controlInterface))
	r.Put("/changeInterval", put.New(controlInterface))
	r.Delete("/deletePinger", handDelete.New(controlInterface))

	slog.Info("starting server", "addr", cfg.Port)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 4 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverError := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverError <- err
		}
	}()

	select {
	case err := <-serverError:
		if err != nil {
			slog.Error("failed to start server", "err", err)
		}
	case <-appCtx.Done():
		slog.Info("Received a signal. Trying graceful shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			_ = srv.Close()
			slog.Error("Could not stop server gracefully", "err", err)
		}
	}

	slog.Info("server stopped")
}
