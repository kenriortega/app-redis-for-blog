package domain

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var (
	INDEX = "pizza:index"
)

func CreateIndexRedisSearch(ctx context.Context, rdb *redis.Client) {

	indices, err := rdb.Do(ctx, `FT._LIST`).Result()
	if err != nil {
		log.Fatal(err)
	}
	for _, index := range indices.([]interface{}) {
		if index.(string) == INDEX {
			log.Println("Find index to drop")
			rdb.Do(ctx, `FT.DROPINDEX`, INDEX)
			break
		}
	}
	rdb.Do(
		ctx,
		`FT.CREATE`, INDEX,
		"ON", "hash",
		"PREFIX", 1, "pizza",
		"SCHEMA",
		"country", "TAG",
		"currency", "TAG",
	)
}
