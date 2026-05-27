package config

import (
	"context"
	"log"
    "github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg *Config) *redis.Client{
	rdb:=redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	if err := rdb.Ping(context.Background()).Err(); err!=nil{
	  log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Redis connected")
	return rdb
}