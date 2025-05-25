package queue

import (
	"context"
	"encoding/json"
	"time"

	"premium_cars_app/pkg/models"

	"github.com/redis/go-redis/v9"
)

const (
	QueueKey     = "request_queue"
	StatusPrefix = "request_status:"
)

type Queue struct {
	rdb *redis.Client
}

func NewQueue(rdb *redis.Client) *Queue {
	return &Queue{rdb: rdb}
}

func (q *Queue) Enqueue(ctx context.Context, req *models.Request) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if err := q.rdb.RPush(ctx, QueueKey, data).Err(); err != nil {
		return err
	}
	return q.rdb.Set(ctx, StatusPrefix+req.ID, "queued", 24*time.Hour).Err()
}

func (q *Queue) GetStatus(ctx context.Context, id string) (string, error) {
	return q.rdb.Get(ctx, StatusPrefix+id).Result()
}

func (q *Queue) Dequeue(ctx context.Context) (*models.Request, error) {
	data, err := q.rdb.LPop(ctx, QueueKey).Result()
	if err != nil {
		return nil, err
	}

	var req models.Request
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (q *Queue) SetStatus(ctx context.Context, id string, status string) error {
	return q.rdb.Set(ctx, StatusPrefix+id, status, 24*time.Hour).Err()
}
