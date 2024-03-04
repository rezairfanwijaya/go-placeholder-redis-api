package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func getRedisClient() *redis.Client {
	env, err := godotenv.Read(".env")
	if err != nil {
		log.Fatalf("failed get env file, err: %s", err)
	}

	return redis.NewClient(&redis.Options{
		Addr: env["REDIS_URL"],
		DB:   0,
	})
}
