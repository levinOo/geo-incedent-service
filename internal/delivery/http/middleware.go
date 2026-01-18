package myHttp

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/levinOo/geo-incedent-service/config"
)

// Middleware для проверки API ключа
func ApiKeyMiddleware(cfg *config.HTTPServerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		if apiKey == "" {
			slog.Error("missing API key header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-API-Key header required"})
			c.Abort()
			return
		}

		if apiKey != cfg.APIKey {
			slog.Error("invalid API key")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		slog.Debug("API key validated")
		c.Next()
	}
}

// Middleware для логирования запросов
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Info("request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("ip", c.ClientIP()),
		)
		c.Next()
	}
}
