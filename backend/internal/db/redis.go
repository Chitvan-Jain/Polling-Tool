package db

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(addr string) *redis.Client {
	client := redis.NewClient(&redis.Options{Addr: addr})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}
	log.Println("connected to Redis")
	return client
}