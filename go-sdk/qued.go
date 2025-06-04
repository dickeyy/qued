package qued

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Message is the structure for queued items.
type Message struct {
	ID        string `json:"id"`
	Type      string `json:"type,omitempty"`
	Payload   any    `json:"payload"`
	CreatedAt string `json:"created_at"`
	Attempts  int    `json:"attempts,omitempty"`
}

// Queue wraps Redis connection and queue name.
type Queue struct {
	client   *redis.Client
	name     string
	deadName string
	maxTries int
}

// NewQueue creates a new Queue instance.
func NewQueue(name, redisURL string, maxTries int) *Queue {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(fmt.Errorf("invalid Redis URL: %w", err))
	}
	return &Queue{
		client:   redis.NewClient(opts),
		name:     name,
		deadName: fmt.Sprintf("%s:dead", name),
		maxTries: maxTries,
	}
}

// Enqueue ads a message to the queue.
func (q *Queue) Enqueue(ctx context.Context, msgType string, payload any) (string, error) {
	msg := Message{
		ID:        uuid.New().String(),
		Type:      msgType,
		Payload:   payload,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Attempts:  0,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	if err := q.client.LPush(ctx, q.name, data).Err(); err != nil {
		return "", err
	}

	return msg.ID, nil
}

// Dequeue retrieves a message from the queue (blocking).
func (q *Queue) Dequeue(ctx context.Context, timeout time.Duration) (*Message, error) {
	res, err := q.client.BRPop(ctx, timeout, q.name).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal([]byte(res[1]), &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// Fail moves a message to the dead-letter queue.
func (q *Queue) Fail(ctx context.Context, msg *Message) error {
	msg.Attempts++
	data, err := json.Marshal(msg)

	if err != nil {
		return err
	}

	return q.client.LPush(ctx, q.deadName, data).Err()
}

// Retry re-enqueues a message, or moves to dead-letter if maxTries exceeded.
func (q *Queue) Retry(ctx context.Context, msg *Message) error {
	msg.Attempts++
	if msg.Attempts >= q.maxTries {
		return q.Fail(ctx, msg)
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return q.client.LPush(ctx, q.name, data).Err()
}

// DeadLetter retrieves a message from the dead-letter queue (blocking).
func (q *Queue) DeadLetter(ctx context.Context, timeout time.Duration) (*Message, error) {
	res, err := q.client.BRPop(ctx, timeout, q.deadName).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal([]byte(res[1]), &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// Close the queue and release resources.
func (q *Queue) Close() error {
	return q.client.Close()
}
