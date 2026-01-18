package worker

import (
	"context"
	"log/slog"

	"github.com/levinOo/geo-incedent-service/config"
	"github.com/levinOo/geo-incedent-service/internal/entity"
	"github.com/levinOo/geo-incedent-service/internal/queue"
	"github.com/levinOo/geo-incedent-service/internal/service"
)

type Worker struct {
	queue  *queue.Queue
	sender *service.WebhookSender
	config config.Worker
}

func NewWorker(queue *queue.Queue, sender *service.WebhookSender, config *config.Config) *Worker {
	return &Worker{
		queue:  queue,
		sender: sender,
		config: config.Worker,
	}
}

func (w *Worker) Run(ctx context.Context, webhookURL string) {

	for {
		select {
		case <-ctx.Done():
			slog.Info("воркер остановлен")
			return
		default:
			task, err := w.queue.Dequeue(ctx)
			if err != nil {
				slog.Error("не удалось получить задачу", "error", err)
				continue
			}

			w.processTask(ctx, task)
		}
	}
}

func (w *Worker) processTask(ctx context.Context, task *entity.WebhookTask) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("паника в воркере", "recover", r)
		}
	}()

	err := w.sender.Send(ctx, task, w.config.WebhookURL)
	if err == nil {
		slog.Info("вебхук успешно отправлен", "id", task.ID)
		w.queue.Ack(ctx, task.ID)
		return
	}

	task.RetryCount++

	if task.RetryCount > w.config.MaxRetries {
		slog.Error("превышено количество попыток, перенос в DLQ", "id", task.ID, "error", err)
		if err := w.queue.MoveToDLQ(ctx, task); err != nil {
			slog.Error("не удалось перенести в DLQ", "error", err)
		}
		return
	}

	slog.Warn("не удалось отправить вебхук, повторная попытка", "id", task.ID, "count", task.RetryCount, "error", err)

	if err := w.queue.Update(ctx, task); err != nil {
		slog.Error("не удалось обновить счетчик попыток", "error", err)
	}

	if err := w.queue.Enqueue(ctx, task); err != nil {
		slog.Error("не удалось вернуть задачу в очередь", "error", err)
	}
}
