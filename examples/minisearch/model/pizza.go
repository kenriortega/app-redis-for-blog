package model

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
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

func NewPizzaR(line []string) PizzaRestaurant {
	dateToParse := strings.Split(line[1], "T")[0]
	date, err := time.Parse("2006-01-02", dateToParse)
	if err != nil {
		log.Fatal(err)
	}
	return PizzaRestaurant{
		ID:              line[0],
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
