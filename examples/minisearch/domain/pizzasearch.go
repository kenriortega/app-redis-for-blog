package domain

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var (
	INDEX = "pizzasIdx"
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
		"description", "TEXT",
		"page_url", "TEXT",
		"category", "TEXT",
		"primary_category", "TEXT",
		"location", "GEO",
		"date_added", "NUMERIC",
		"country", "TAG",
		"currency", "TAG",
	)
}

func CreateIndexJSONRedisSearch(ctx context.Context, rdb *redis.Client) {

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
	_, err = rdb.Do(
		ctx,
		`FT.CREATE`, INDEX,
		"ON", "JSON", "SCHEMA",
		"$.name", "AS", "name", "TEXT",
		"$.date_added", "AS", "date_added", "NUMERIC",
		"$.category", "AS", "category", "TEXT",
		"$.primary_category", "AS", "subcategory", "TEXT",
		"$.location", "as", "loc", "GEO",
	).Result()
	if err != nil {
		log.Fatal(err)
	}
}
