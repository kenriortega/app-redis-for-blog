package domain

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
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

func NewPizzaR(line []string, index int) PizzaRestaurant {
	dateToParse := strings.Split(line[1], "T")[0]
	date, err := time.Parse("2006-01-02", dateToParse)
	if err != nil {
		log.Fatal(err)
	}
	return PizzaRestaurant{
		ID:              fmt.Sprintf(`%s_%d`, line[0], index),
		DateAdded:       fmt.Sprintf("%d", date.Unix()),
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

// IngestData excute a simple task to load data from csv pizza to redis
func IngestData(
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

		pizzaR := NewPizzaR(line, index)

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
