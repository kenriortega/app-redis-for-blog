package main

import (
	"app/pkg/db"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
func drainBody(respBody io.ReadCloser) {
	// Callers should close resp.Body when done reading from it.
	// If resp.Body is not closed, the Client's underlying RoundTripper
	// (typically Transport) may not be able to re-use a persistent TCP
	// connection to the server for a subsequent "keep-alive" request.
	if respBody != nil {
		// Drain any remaining Body and then close the connection.
		// Without this closing connection would disallow re-using
		// the same connection for future uses.
		//  - http://stackoverflow.com/a/17961593/4465767
		defer respBody.Close()
		_, _ = io.Copy(ioutil.Discard, respBody)
	}
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
	defer drainBody(res.Body)

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

func main() {
	// Create redis instance
	rdb := db.GetRedisDbClient(context.Background())

	url := fmt.Sprintf("%s/%s/%s", urlBase, urlVersion, resourceCoinList)
	fmt.Println("Fetching all coins from: ", url)

	start := time.Now()
	resp, err := getCoins(
		context.Background(),
		rdb,
		"GET",
		url,
		"coins:list",
		60*time.Second,
	)
	if err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)
	resp.ResponseTime = elapsed.String()

	fmt.Printf("Fetched [%d] coins from source: %s response time: %s\n", len(resp.Coins), resp.Source, resp.ResponseTime)
}
