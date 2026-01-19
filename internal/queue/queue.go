package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/levinOo/geo-incedent-service/internal/entity"
	"github.com/redis/go-redis/v9"
)

type Queue struct {
	client *redis.Client
}

func NewQueue(client *redis.Client) *Queue {
	return &Queue{client: client}
}

func (q *Queue) Enqueue(ctx context.Context, task *entity.WebhookTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		slog.Error("не удалось сериализовать webhook", "error", err)
		return fmt.Errorf("не удалось сериализовать webhook: %w", err)
	}

	taskKey := fmt.Sprintf("webhook:task:%s", task.ID)

	pipe := q.client.Pipeline()
	pipe.Set(ctx, taskKey, data, 24*time.Hour)
	pipe.LPush(ctx, "webhook:pending", task.ID.String())

	_, err = pipe.Exec(ctx)
	if err != nil {
		slog.Error("не удалось добавить задачу в очередь", "error", err)
		return fmt.Errorf("не удалось добавить задачу в очередь: %w", err)
	}

	return nil
}

func (q *Queue) Dequeue(ctx context.Context) (*entity.WebhookTask, error) {
	result, err := q.client.BRPop(ctx, 0, "webhook:pending").Result()
	if err != nil {
		slog.Error("не удалось получить задачу из очереди", "error", err)
		return nil, fmt.Errorf("не удалось получить задачу из очереди: %w", err)
	}

	taskID := result[1]

	taskKey := fmt.Sprintf("webhook:task:%s", taskID)

	data, err := q.client.Get(ctx, taskKey).Bytes()
	if err != nil {
		slog.Error("не удалось получить задачу из очереди", "error", err)
		return nil, fmt.Errorf("не удалось получить задачу из очереди: %w", err)
	}

	var task entity.WebhookTask
	if err := json.Unmarshal(data, &task); err != nil {
		slog.Error("не удалось десериализовать webhook", "error", err)
		return nil, fmt.Errorf("не удалось десериализовать webhook: %w", err)
	}

	return &task, nil
}

func (q *Queue) Ack(ctx context.Context, taskID uuid.UUID) error {
	taskKey := fmt.Sprintf("webhook:task:%s", taskID)

	return q.client.Del(ctx, taskKey).Err()
}

func (q *Queue) Update(ctx context.Context, task *entity.WebhookTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		slog.Error("не удалось сериализовать webhook", "error", err)
		return fmt.Errorf("не удалось сериализовать webhook: %w", err)
	}

	taskKey := fmt.Sprintf("webhook:task:%s", task.ID)
	return q.client.Set(ctx, taskKey, data, 24*time.Hour).Err()
}

func (q *Queue) MoveToDLQ(ctx context.Context, task *entity.WebhookTask) error {
	taskKey := fmt.Sprintf("webhook:task:%s", task.ID)

	pipe := q.client.Pipeline()
	pipe.LPush(ctx, "webhook:queue:dead", task.ID.String())
	pipe.Persist(ctx, taskKey)
	_, err := pipe.Exec(ctx)
	if err != nil {
		slog.Error("не удалось переместить задачу в очередь", "error", err)
		return fmt.Errorf("не удалось переместить задачу в очередь: %w", err)
	}
	return nil
}
