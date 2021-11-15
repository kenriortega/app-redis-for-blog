package main

import (
	"app/pkg/db"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"app/examples/minisearch/model"

	"github.com/go-redis/redis/v8"
)

var (
	action string
)

func init() {
	// cmds
	flag.StringVar(&action, "action", "seed", "seed pizza data on redis")
	flag.Parse()
}
func main() {
	// defined variables for our tasks
	ctx := context.Background()
	rdb := db.GetRedisDbClient(context.Background())
	// cmd managements
	switch action {

	case "seed":
		start := time.Now()
		var path, _ = os.Getwd()
		// ingest data from csv to redis using pipeline cmd
		ingestDataToRedis(
			ctx,
			rdb,
			path,
			"data/pizzas.csv",
		)
		elapsed := time.Since(start)
		log.Printf("Seed pizza data on redis [%s]\n", elapsed.String())
	}

}

// ingestDataToRedis return and array of []PizzaRestaurant parsed from CSV
func ingestDataToRedis(
	ctx context.Context,
	rdb *redis.Client,
	path, filename string,
) {
	pipe := rdb.Pipeline()

	csvFile, err := os.Open(fmt.Sprintf("%s/%s", path, filename))
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	for index, line := range csvLines[1:] {

		pizzaR := model.NewPizzaR(line)

		key := fmt.Sprintf(`pizza:%s_%d`, pizzaR.ID, index)
		value, err := pizzaR.ToMAP()
		if err != nil {
			log.Fatal(err)
		}
		pipe.HSet(ctx, key, value)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		panic(err)
	}
	log.Println("Successfully Ingested CSV file on redis")
}
