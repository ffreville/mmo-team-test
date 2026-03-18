package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisClient(host string, port int, db int, password string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to ping redis: %w", err)
	}

	log.Println("Redis connection established successfully")

	return &RedisClient{client: client, ctx: ctx}, nil
}

func (r *RedisClient) Close() {
	r.client.Close()
	log.Println("Redis connection closed")
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}

func (r *RedisClient) SetSession(userID string, token string, expiry time.Duration) error {
	key := fmt.Sprintf("session:%s", userID)
	return r.client.Set(r.ctx, key, token, expiry).Err()
}

func (r *RedisClient) GetSession(userID string) (string, error) {
	key := fmt.Sprintf("session:%s", userID)
	return r.client.Get(r.ctx, key).Result()
}

func (r *RedisClient) DeleteSession(userID string) error {
	key := fmt.Sprintf("session:%s", userID)
	return r.client.Del(r.ctx, key).Err()
}

func (r *RedisClient) SetRateLimit(key string, limit int, window time.Duration) (bool, error) {
	current, err := r.client.Incr(r.ctx, key).Result()
	if err != nil {
		return false, err
	}

	if current == 1 {
		if err := r.client.Expire(r.ctx, key, window).Err(); err != nil {
			return false, err
		}
	}

	return current <= int64(limit), nil
}
