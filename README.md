# Qued

A minimal, language-agnostic, Redis-backed queue library.

## Purpose

I wanted a simple queue system that could be used in different programs concurrently. Through my research, I found that anything that was similar to what I was looking for was either too complex, too expensive, or didn't support multiple languages.

With Qued, you can create a queue, add items to it, and consume them from any number of workers in any language supported. You bring your own Redis instance, and Qued will handle the rest (as far as providing queue/dequeue functionality).

## What it is

Qued is a Redis-backed queue library that is designed to be simple, easy to use, and secure. It is made to be used in any supported language, and any environment. It was made to be used to handle events in a microservices architecture.

Qued is not a full-featured queue system. It is not fully complete, and it is not a replacement for cloud-based message brokers. If you are looking for a simple way to handle events across services, and ensure data-privacy, then Qued is for you.

## How it works

1. Provide a connection string and a queue name, Qued will create a Redis list with that name.
2. Enqueue some data, Qued will encrypt the data and add it to the list using `LPUSH`.
3. Dequeue some data when you are ready to process it, Qued will decrypt the data and return it to you, in the background it will remove the item from the list.

Qued will automatically requeue an item if something goes wrong.

## Example

#### Producer

```ts
import { Qued } from "qued";

const queue = new Qued({
  name: "my-queue",
  encryptionKey: "my-encryption-key", // optional but recommended
  redis: {
    connectionString: "redis://localhost:6379",
  },
});

const data = {
  name: "John Doe",
  email: "john.doe@example.com",
};

const { result, error } = await queue.enqueue(data, {
  ttl: 1000 * 60 * 60 * 24, // optional, the time to live for the item in milliseconds
});

if (error) {
  // handle error
}
```

#### Consumer

```go
import (
    "github.com/dickeyy/qued"
)

func main() {
    queue := qued.NewQueue(
        "my-queue",
        "my-encryption-key", // optional but recommended
        "redis://localhost:6379",
    )

    for {
        data, err := queue.Dequeue()
        if err != nil {
            // handle error
        }

        // do something with the data
    }
}
```

Producers and consumers can be in the same language, or different languages. The SDKs provide both methods for each language.

More examples can be found in the [examples](./examples) directory.

## Features

- Language-agnostic
- Bring your own Redis (or some Redis API compatible DB)
- Simple, easy to use
- Data-agnostic (the structure of the data is up to you)
- Secure (Qued encrypts and decrypts the data for you)

## Roadmap

- [ ] Add a way to tell the producer that the consumer has processed the item
- [ ] Dead-letter queues
- [ ] Scheduled processing (tell the consumer when to process an item, or how long to wait before processing an item)
