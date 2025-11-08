package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var CacheEnabled bool

func ConnectRedis() {
	CacheEnabled = os.Getenv("CACHE_ENABLED") == "true"

	if !CacheEnabled {
		log.Println("Redis cache is disabled")
		return
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
		}
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       redisDB,
		PoolSize: 10,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v. Cache will be disabled.", err)
		CacheEnabled = false
		return
	}

	log.Println("Redis cache connected successfully!")
}

func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// CacheGet retrieves a value from cache
func CacheGet(ctx context.Context, key string, dest interface{}) error {
	if !CacheEnabled || RedisClient == nil {
		return redis.Nil
	}

	val, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// CacheSet stores a value in cache with TTL
func CacheSet(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !CacheEnabled || RedisClient == nil {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return RedisClient.Set(ctx, key, data, ttl).Err()
}

// CacheDelete removes a key from cache
func CacheDelete(ctx context.Context, key string) error {
	if !CacheEnabled || RedisClient == nil {
		return nil
	}

	return RedisClient.Del(ctx, key).Err()
}

// CacheDeletePattern removes all keys matching a pattern
func CacheDeletePattern(ctx context.Context, pattern string) error {
	if !CacheEnabled || RedisClient == nil {
		return nil
	}

	keys, err := RedisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return RedisClient.Del(ctx, keys...).Err()
	}

	return nil
}

// GetCacheTTL returns the default cache TTL from environment
func GetCacheTTL() time.Duration {
	ttlStr := os.Getenv("CACHE_TTL")
	if ttlStr == "" {
		return 5 * time.Minute // Default 5 minutes
	}

	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		return 5 * time.Minute
	}

	return time.Duration(ttl) * time.Second
}
