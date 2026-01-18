package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("ошибка чтения тела запроса", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var payload interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			slog.Error("ошибка парсинга JSON", "error", err, "body", string(body))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		slog.Info("получен вебхук", "payload", payload, "headers", r.Header)
		w.WriteHeader(http.StatusOK)
	})

	slog.Info("Webhook Stub запущен на порту :9090")
	if err := http.ListenAndServe(":9090", nil); err != nil {
		slog.Error("ошибка запуска сервера", "error", err)
		os.Exit(1)
	}
}
