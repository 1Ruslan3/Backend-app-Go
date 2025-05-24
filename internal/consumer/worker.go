package consumer

import (
	"context"
	"log"
	"premium_cars_app/internal/queue"
	"time"
)

type Worker struct {
	Queue *queue.Queue
}

func (w *Worker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			req, err := w.Queue.Dequeue(ctx)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			log.Printf("Processing request %s for user %s", req.ID, req.UserID)

			// Симулируем обработку
			time.Sleep(2 * time.Second)

			if err := w.Queue.SetStatus(ctx, req.ID, "processed"); err != nil {
				log.Printf("Failed to update status: %v", err)
			}

			log.Printf("Request %s processed", req.ID)
		}
	}
}
