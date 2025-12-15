package config

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

// Ctx is the shared context for Redis operations
var Ctx = context.Background()

// Rdb is the global Redis client
var Rdb *redis.Client

// InitEnv loads environment variables from .env file
func InitEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
}

// InitRedis initializes the Redis connection
func InitRedis() {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		log.Println("ℹ️ Redis not configured; starting without clustering")
		return
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	password := os.Getenv("REDIS_PASSWORD")
	username := os.Getenv("REDIS_USERNAME")

	Rdb = redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Username: username,
		Password: password,
		DB:       0,
	})

	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		log.Printf("⚠️ Could not connect to Redis: %v (continuing without clustering)", err)
		Rdb = nil
		return
	}
	log.Println("✅ Successfully connected to Redis for WebRTC signaling")
}
