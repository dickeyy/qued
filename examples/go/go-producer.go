package main

import (
	"context"
	"fmt"
	"log"

	qued "github.com/dickeyy/qued/go-sdk"
)

func producer() {
	ctx := context.Background()
	queue := qued.NewQueue("my-queue", "redis://localhost:6379", 3)

	id, err := queue.Enqueue(ctx, "user.created", map[string]string{
		"name":  "John Doe",
		"email": "john.doe@example.com",
	})
	if err != nil {
		log.Fatalf("Enqueue failed: %v", err)
	}

	fmt.Printf("Enqueued message with ID: %s\n", id)
}
