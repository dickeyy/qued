import { createClient, type RedisClientType } from "redis";
import { v4 as uuidv4 } from "uuid";

export interface QuedOptions {
  name: string;
  redisUrl: string;
  maxTries?: number;
}

export interface QuedMessage<T = any> {
  id: string;
  type?: string;
  payload: T;
  created_at: string;
  attempts: number;
}

export class Qued {
  private name: string;
  private deadName: string;
  private maxTries: number;
  private client: RedisClientType;

  constructor({ name, redisUrl, maxTries = 3 }: QuedOptions) {
    this.name = name;
    this.deadName = `${name}:dead`;
    this.maxTries = maxTries;
    this.client = createClient({ url: redisUrl });
    this.client.connect();
  }

  /**
   * Enqueue a message to the queue.
   * @param type Message type (optional)
   * @param payload Message payload
   * @returns The message ID
   */
  async enqueue<T = any>(
    type: string | undefined,
    payload: T
  ): Promise<string> {
    const msg: QuedMessage<T> = {
      id: uuidv4(),
      type,
      payload,
      created_at: new Date().toISOString(),
      attempts: 0,
    };

    await this.client.lPush(this.name, JSON.stringify(msg));
    return msg.id;
  }

  /**
   * Dequeue a message from the queue (blocking pop).
   * @param timeout Timeout in seconds (0 = block forever)
   * @returns The message or null if timeout
   */
  async dequeue<T = any>(timeout = 0): Promise<QuedMessage<T> | null> {
    // BRPOP returns { key, element } or null
    const result = await this.client.brPop(this.name, timeout);
    if (!result) return null;
    return JSON.parse(result.element) as QuedMessage<T>;
  }

  /**
   * Move a message to the dead-letter queue.
   * @param msg The message to move
   */
  async fail<T = any>(msg: QuedMessage<T>): Promise<void> {
    msg.attempts = (msg.attempts || 0) + 1;
    await this.client.lPush(this.deadName, JSON.stringify(msg));
  }

  /**
   * Retry a message, or move to dead-letter if maxTries exceeded.
   * @param msg The message to retry
   */
  async retry<T = any>(msg: QuedMessage<T>): Promise<void> {
    msg.attempts = (msg.attempts || 0) + 1;
    if (msg.attempts >= this.maxTries) {
      await this.fail(msg);
    } else {
      await this.client.lPush(this.name, JSON.stringify(msg));
    }
  }

  /**
   * Dequeue a message from the dead-letter queue (blocking pop).
   * @param timeout Timeout in seconds (0 = block forever)
   * @returns The message or null if timeout
   */
  async deadLetter<T = any>(timeout = 0): Promise<QuedMessage<T> | null> {
    const result = await this.client.brPop(this.deadName, timeout);
    if (!result) return null;
    return JSON.parse(result.element) as QuedMessage<T>;
  }

  /**
   * Disconnect from Redis.
   */
  async disconnect(): Promise<void> {
    await this.client.close();
  }
}
