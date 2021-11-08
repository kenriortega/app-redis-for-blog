package db

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

//Used to execute client creation procedure only once.
var redisOnce sync.Once

// Create redis instance
func GetRedisDbClient(ctx context.Context) *redis.Client {
	var clientInstance *redis.Client
	redisOnce.Do(func() {

		client := redis.NewClient(&redis.Options{
			Addr:         os.Getenv("REDIS_URI"),
			Username:     "",
			Password:     os.Getenv("REDIS_PASS"),
			DB:           0,
			DialTimeout:  60 * time.Second,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
		})

		_, err := client.Ping(context.TODO()).Result()
		if err != nil {
			log.Fatal(err)
		}
		clientInstance = client
	})

	return clientInstance
}
