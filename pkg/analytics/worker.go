package analytics

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const streamName = "url:clicks"

type Worker struct {
	redisClient *redis.Client
	onClick     func(code string) error
	onCleanup   func() error
}

func NewWorker(r *redis.Client, onClick func(code string) error, onCleanup func() error) *Worker {
	return &Worker{
		redisClient: r,
		onClick:     onClick,
		onCleanup:   onCleanup,
	}
}

func (w *Worker) PublishClick(code string) error {
	return w.redisClient.XAdd(context.Background(), &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{
			"short_code": code,
			"timestamp":  time.Now().Unix(),
		},
	}).Err()
}

func (w *Worker) Run(ctx context.Context) {
	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("analytics worker shutting down")
			return

		case <-cleanupTicker.C:
			if err := w.onCleanup(); err != nil {
				log.Printf("failed to cleanup expired URLs: %v", err)
			} else {
				log.Println("expired URLs cleaned up successfully")
			}

		default:
			streams, err := w.redisClient.XRead(ctx, &redis.XReadArgs{
				Streams: []string{streamName, "0"},
				Count:   100,
				Block:   5 * time.Second,
			}).Result()
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					code, ok := msg.Values["short_code"].(string)
					if !ok {
						w.redisClient.XDel(ctx, streamName, msg.ID)
						continue
					}
					if err := w.onClick(code); err != nil {
						log.Printf("failed to process click for %s: %v", code, err)
						continue
					}
					w.redisClient.XDel(ctx, streamName, msg.ID)
				}
			}
		}
	}
}
