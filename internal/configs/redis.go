package configs

import (
	"context"
	"os"
	"github.com/redis/go-redis/v9"
	"log"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis() {
	addr := os.Getenv("REDIS_ADDR")
	pass := os.Getenv("REDIS_PASSWORD")

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       0,
	})

	// Test koneksi
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Gagal terhubung ke Redis: %v", err)
	}

	log.Println("Koneksi ke Redis Berhasil!")
}