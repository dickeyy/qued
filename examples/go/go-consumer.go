package main

import (
	"context"

	qued "github.com/dickeyy/qued/go-sdk"
)

func consumer() {
	ctx := context.Background()
	queue := qued.NewQueue("my-queue", "redis://localhost:6379", 3)

	for {
		msg, err := queue.Dequeue(ctx, 0)
		if err != nil {
			println("Dequeue failed: ", err)
			continue
		}

		if msg == nil {
			continue
		}

		// Do something with the message
		// err := processMessage(msg)

		// If something goes wrong, retry the message
		if err := queue.Retry(ctx, msg); err != nil {
			println("Retry failed: ", err)
		}
	}
}
