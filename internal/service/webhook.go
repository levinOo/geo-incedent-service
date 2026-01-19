package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/levinOo/geo-incedent-service/config"
	"github.com/levinOo/geo-incedent-service/internal/entity"
)

type WebhookSender struct {
	client *retryablehttp.Client
}

func NewWebhookSender(cfg *config.Config) *WebhookSender {
	retryClient := retryablehttp.NewClient()

	retryClient.RetryMax = cfg.RetryClient.RetryMax
	retryClient.RetryWaitMin = cfg.RetryClient.RetryWaitMin
	retryClient.RetryWaitMax = cfg.RetryClient.RetryWaitMax
	retryClient.HTTPClient.Timeout = cfg.RetryClient.Timeout

	retryClient.Logger = nil

	return &WebhookSender{
		client: retryClient,
	}
}

func (s *WebhookSender) Send(ctx context.Context, task *entity.WebhookTask, webhookURL string) error {
	payload := entity.WebhookPayload{
		Name:       task.Name,
		IncidentID: task.IncidentID,
		UserID:     task.UserID,
		Timestamp:  time.Now().UTC(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, webhookURL, data)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("вебхук вернул неверный статус: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}
