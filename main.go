package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	urlBase          = "https://api.coingecko.com/api"
	urlVersion       = "v3"
	resourceCoinList = "coins/list"
)

// Coin ...
type Coin struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}
type ResponseCoins struct {
	Coins        []Coin `json:"coins,omitempty"`
	Source       string `json:"soure,omitempty"`
	ResponseTime string `json:"response_time,omitempty"`
}

// ToJSON ...
func (r *ResponseCoins) ToJSON() string {
	bytes, err := json.Marshal(r)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return string(bytes)
}

func main() {
	// Create redis instance
	rdb := GetRedisDbClient(context.Background())

	url := fmt.Sprintf("%s/%s/%s", urlBase, urlVersion, resourceCoinList)
	fmt.Println("Fetching all coins from: ", url)

	start := time.Now()
	resp, err := getCoins(
		context.Background(),
		rdb,
		"GET",
		url,
		"coins:list",
		10*time.Second,
	)
	if err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)
	resp.ResponseTime = elapsed.String()

	fmt.Printf("Fetched [%d] coins from source: %s response time: %s\n", len(resp.Coins), resp.Source, resp.ResponseTime)
}

// getCoins ...
func getCoins(
	ctx context.Context,
	rdb *redis.Client,
	method, endpoint, key string,
	duration time.Duration,
) (responeCoins ResponseCoins, err error) {

	result, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Println("key not found:", err)
	} else if err != nil {
		return
	}

	if result != "" {
		err = json.Unmarshal([]byte(result), &responeCoins)
		if err != nil {
			return
		}
		responeCoins.Source = "cache"
		return responeCoins, nil
	}

	client := &http.Client{}
	requestUrl, err := url.Parse(endpoint)
	if err != nil {
		return
	}
	req, err := http.NewRequest(method, requestUrl.String(), nil)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	coins := []Coin{}
	if err = json.NewDecoder(res.Body).Decode(&coins); err != nil {
		return
	}
	responeCoins.Coins = coins
	responeCoins.Source = "API"
	err = rdb.Set(ctx, key, responeCoins.ToJSON(), duration).Err()
	if err != nil {
		return
	}
	return
}

// Create redis instance
func GetRedisDbClient(ctx context.Context) *redis.Client {

	clientInstance := redis.NewClient(&redis.Options{
		Addr:         os.Getenv("REDIS_URI"),
		Username:     "",
		Password:     os.Getenv("REDIS_PASS"),
		DB:           0,
		DialTimeout:  60 * time.Second,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	})

	_, err := clientInstance.Ping(context.TODO()).Result()
	if err != nil {
		log.Fatal(err)
	}

	return clientInstance
}
