package main

import (
	"app/pkg/db"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

// PizzaRestaurant ...
type PizzaRestaurant struct {
	ID              string `json:"id,omitempty"`
	DateAdded       string `json:"date_added,omitempty"`
	Address         string `json:"address,omitempty"`
	Category        string `json:"category,omitempty"`
	PrimaryCategory string `json:"primary_category,omitempty"`
	City            string `json:"city,omitempty"`
	Country         string `json:"country,omitempty"`
	Location        string `json:"location,omitempty"`
	PageURL         string `json:"page_url,omitempty"`
	AmmountMAX      string `json:"ammount_max,omitempty"`
	Currency        string `json:"currency,omitempty"`
	Description     string `json:"description,omitempty"`
	Name            string `json:"name,omitempty"`
}

// ToJSON ...
func (p *PizzaRestaurant) ToJSON() string {
	bytes, err := json.Marshal(p)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return string(bytes)
}

// ToMAP ...
func (p *PizzaRestaurant) ToMAP() (toHashMap map[string]interface{}, err error) {

	fromStruct, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(fromStruct, &toHashMap); err != nil {
		return toHashMap, err
	}

	return toHashMap, nil
}

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
	for _, line := range csvLines {
		pizzaR := PizzaRestaurant{
			ID:              line[0],
			DateAdded:       line[1],
			Address:         line[3],
			Category:        line[4],
			PrimaryCategory: line[5],
			City:            line[6],
			Country:         line[7],
			// redis geo types are declared "longitude,latitude"
			Location:    fmt.Sprintf("%s,%s", line[10], line[9]),
			PageURL:     line[11],
			AmmountMAX:  line[12],
			Currency:    line[14],
			Description: line[16],
			Name:        line[17],
		}

		// Redis Insgestion Proccess
		key := fmt.Sprintf(`pizza:%s`, pizzaR.ID)
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
