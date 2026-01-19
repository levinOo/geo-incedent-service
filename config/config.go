package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	HTTPServer  HTTPServerConfig
	Postgres    PostgresConfig
	Redis       RedisConfig
	RetryClient RetryClient
	Worker      Worker
}

type HTTPServerConfig struct {
	HTTPServerPort         string
	HTTPServerReadTimeout  time.Duration
	HTTPServerWriteTimeout time.Duration
	HTTPServerIdleTimeout  time.Duration
	APIKey                 string

	StatsWindowMinutes int
}

type PostgresConfig struct {
	PostgresURL                         string
	PostgresMaxConns                    int
	PostgresMinConns                    int
	PostgresMaxConnLifetime             time.Duration
	PostgresConnectTimeout              time.Duration
	PostgresMaxConnWaitTimeout          time.Duration
	PostgresMaxCachedPreparedStatements int
	PostgresPreferSimpleProtocol        bool
}

type RedisConfig struct {
	RedisAddr         string
	RedisPoolSize     int
	RedisMinIdleConns int
	RedisPoolTimeout  time.Duration
	RedisDialTimeout  time.Duration
	RedisReadTimeout  time.Duration
	RedisWriteTimeout time.Duration
	RedisMaxRetries   int
}

type Worker struct {
	WebhookURL string
	MaxRetries int
}

type RetryClient struct {
	RetryMax     int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
	Timeout      time.Duration
}

func Load() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Warn("конфиг файл не найден, использую только переменные окружения")
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	required := []string{"DATABASE_URL", "API_KEY", "WEBHOOK_URL"}

	for _, key := range required {
		if viper.GetString(key) == "" {
			return nil, fmt.Errorf("неустановлено обязательное окружение %s", key)
		}
	}

	cfg := &Config{
		HTTPServer: HTTPServerConfig{
			HTTPServerPort:         mustLoad("HTTP_SERVER_PORT"),
			HTTPServerReadTimeout:  viper.GetDuration("HTTP_SERVER_READ_TIMEOUT"),
			HTTPServerWriteTimeout: viper.GetDuration("HTTP_SERVER_WRITE_TIMEOUT"),
			HTTPServerIdleTimeout:  viper.GetDuration("HTTP_SERVER_IDLE_TIMEOUT"),
			APIKey:                 mustLoad("API_KEY"),
			StatsWindowMinutes:     viper.GetInt("STATS_WINDOW_MINUTES"),
		},
		Postgres: PostgresConfig{
			PostgresURL:                         mustLoad("DATABASE_URL"),
			PostgresMaxConns:                    viper.GetInt("DATABASE_MAX_CONNS"),
			PostgresMinConns:                    viper.GetInt("DATABASE_MIN_CONNS"),
			PostgresMaxConnLifetime:             viper.GetDuration("DATABASE_MAX_CONN_LIFETIME"),
			PostgresConnectTimeout:              viper.GetDuration("DATABASE_CONNECT_TIMEOUT"),
			PostgresMaxConnWaitTimeout:          viper.GetDuration("DATABASE_MAX_CONN_WAIT_TIMEOUT"),
			PostgresMaxCachedPreparedStatements: viper.GetInt("DATABASE_MAX_CACHED_PREPARED_STATEMENTS"),
			PostgresPreferSimpleProtocol:        viper.GetBool("DATABASE_PREFER_SIMPLE_PROTOCOL"),
		},
		Redis: RedisConfig{
			RedisAddr:         mustLoad("REDIS_ADDR"),
			RedisPoolSize:     viper.GetInt("REDIS_POOL_SIZE"),
			RedisMinIdleConns: viper.GetInt("REDIS_MIN_IDLE_CONNS"),
			RedisPoolTimeout:  viper.GetDuration("REDIS_POOL_TIMEOUT"),
			RedisDialTimeout:  viper.GetDuration("REDIS_DIAL_TIMEOUT"),
			RedisReadTimeout:  viper.GetDuration("REDIS_READ_TIMEOUT"),
			RedisWriteTimeout: viper.GetDuration("REDIS_WRITE_TIMEOUT"),
			RedisMaxRetries:   viper.GetInt("REDIS_MAX_RETRIES"),
		},
		Worker: Worker{
			WebhookURL: mustLoad("WEBHOOK_URL"),
			MaxRetries: viper.GetInt("WORKER_MAX_RETRIES"),
		},
		RetryClient: RetryClient{
			RetryMax:     viper.GetInt("RETRY_MAX"),
			RetryWaitMin: viper.GetDuration("RETRY_WAIT_MIN"),
			RetryWaitMax: viper.GetDuration("RETRY_WAIT_MAX"),
			Timeout:      viper.GetDuration("RETRY_TIMEOUT"),
		},
	}

	return cfg, nil
}

func mustLoad(name string) string {
	value := viper.GetString(name)
	if value == "" {
		slog.Error("неустановлено обязательное окружение", "error", fmt.Errorf("неустановлено обязательное окружение %s", name))
		os.Exit(1)
	}
	return value
}
