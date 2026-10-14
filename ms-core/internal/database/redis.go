package database

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func ConnectRedis() {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = "localhost:6379" // Default para desarrollo local si no está en docker
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: "", // No password por defecto
		DB:       0,  // DB por defecto
	})

	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis at %s: %v. Cache will be disabled.", redisUrl, err)
		RedisClient = nil
		return
	}

	log.Println("Connected to Redis successfully")
}

func GetRedisClient() *redis.Client {
	return RedisClient
}
