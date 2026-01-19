package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/levinOo/geo-incedent-service/config"
	"github.com/levinOo/geo-incedent-service/internal/db"
	myHttp "github.com/levinOo/geo-incedent-service/internal/delivery/http"
	"github.com/levinOo/geo-incedent-service/internal/queue"
	"github.com/levinOo/geo-incedent-service/internal/repo"
	"github.com/levinOo/geo-incedent-service/internal/service"
	"github.com/levinOo/geo-incedent-service/internal/worker"
)

func Run() error {
	// Context для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Инициализация логгера
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// Конфиг
	cfg, err := config.Load()
	if err != nil {
		slog.Error("ошибка загрузки конфигурации", "error", err)
		return err
	}

	// Postgres
	postgres, err := db.NewPostgres(&cfg.Postgres)
	if err != nil {
		slog.Error("ошибка подключения к postgres", "error", err)
		return err
	}

	slog.Info("postgres подключен")

	// Redis
	redisClient, err := db.NewRedis(&cfg.Redis)
	if err != nil {
		slog.Error("ошибка подключения к redis", "error", err)
		postgres.Pool.Close()
		return err
	}

	slog.Info("redis подключен")

	// Миграции
	if err := db.RunMigrations(postgres); err != nil {
		slog.Error("ошибка выполнения миграций", "error", err)
		return err
	}

	// Слои
	repository := repo.NewRepo(postgres.Pool)
	svc := service.NewService(repository, cfg, redisClient)

	webhookSender := service.NewWebhookSender(cfg)
	q := queue.NewQueue(redisClient.Client)
	bgWorker := worker.NewWorker(q, webhookSender, cfg)

	// Воркер
	go func() {
		slog.Info("воркер запущен")
		bgWorker.Run(ctx)
	}()

	router := myHttp.NewRouter(&cfg.HTTPServer, svc)

	// HTTP Server
	srv := &http.Server{
		Addr:         ":" + cfg.HTTPServer.HTTPServerPort,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.HTTPServerReadTimeout,
		WriteTimeout: cfg.HTTPServer.HTTPServerWriteTimeout,
		IdleTimeout:  cfg.HTTPServer.HTTPServerIdleTimeout,
	}

	// Запуск сервера в горутине
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ошибка запуска http сервера", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("http сервер запущен", "port", cfg.HTTPServer.HTTPServerPort)

	// graceful shutdown
	<-ctx.Done()
	slog.Info("получен сигнал остановки")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("closing database connections...")

	if err := redisClient.Close(); err != nil {
		slog.Error("failed to close redis", "error", err)
	}

	postgres.Pool.Close()

	slog.Info("server exited properly")
	return nil
}
