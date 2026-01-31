package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kexi/mail-to-tg/pkg/config"
)

type Redis struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedis(cfg *config.RedisConfig) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Redis{
		client: client,
		ctx:    ctx,
	}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) Ping() error {
	return r.client.Ping(r.ctx).Err()
}

// Key-value operations
func (r *Redis) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

func (r *Redis) Get(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *Redis) Del(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

func (r *Redis) Exists(key string) (bool, error) {
	n, err := r.client.Exists(r.ctx, key).Result()
	return n > 0, err
}

// Hash operations for reply state
func (r *Redis) HSet(key, field string, value interface{}) error {
	return r.client.HSet(r.ctx, key, field, value).Err()
}

func (r *Redis) HGet(key, field string) (string, error) {
	val, err := r.client.HGet(r.ctx, key, field).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *Redis) HGetAll(key string) (map[string]string, error) {
	return r.client.HGetAll(r.ctx, key).Result()
}

func (r *Redis) HDel(key string, fields ...string) error {
	return r.client.HDel(r.ctx, key, fields...).Err()
}

// List operations for queue
func (r *Redis) LPush(key string, values ...interface{}) error {
	return r.client.LPush(r.ctx, key, values...).Err()
}

func (r *Redis) RPush(key string, values ...interface{}) error {
	return r.client.RPush(r.ctx, key, values...).Err()
}

func (r *Redis) BRPop(timeout time.Duration, keys ...string) ([]string, error) {
	result, err := r.client.BRPop(r.ctx, timeout, keys...).Result()
	if err == redis.Nil {
		return nil, nil
	}
	return result, err
}

func (r *Redis) LLen(key string) (int64, error) {
	return r.client.LLen(r.ctx, key).Result()
}

// Expiration
func (r *Redis) Expire(key string, expiration time.Duration) error {
	return r.client.Expire(r.ctx, key, expiration).Err()
}

// Client getter for advanced operations
func (r *Redis) Client() *redis.Client {
	return r.client
}

func (r *Redis) Context() context.Context {
	return r.ctx
}
