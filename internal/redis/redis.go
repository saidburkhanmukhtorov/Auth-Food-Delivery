package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/time_capsule/Auth-Servic-Timecapsule/config"
)

// Client represents a Redis client.
type Client struct {
	*redis.Client
}

// Connect establishes a connection to the Redis server.
func Connect(cfg *config.Config) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test the connection
	// if _, err := client.Ping(context.Background()).Result(); err != nil {
	// 	return nil, fmt.Errorf("redis connection failed: %w", err)
	// }

	return &Client{client}, nil
}

// SaveOTP saves the OTP code in Redis with an expiration time.
func (c *Client) SaveOTP(ctx context.Context, email string, otp string, expiration time.Duration) error {
	key := fmt.Sprintf("otp:%s", email)
	err := c.Set(ctx, key, otp, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to save OTP in Redis: %w", err)
	}
	return nil
}

// VerifyOTP verifies the OTP code against the one stored in Redis.
func (c *Client) VerifyOTP(ctx context.Context, email string, otp string) (bool, error) {
	key := fmt.Sprintf("otp:%s", email)
	storedOTP, err := c.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil // OTP not found (expired or never set)
		}
		return false, fmt.Errorf("failed to get OTP from Redis: %w", err)
	}

	return storedOTP == otp, nil
}
