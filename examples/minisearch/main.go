package main

import (
	"app/examples/minisearch/model"
	"app/pkg/db"
	"app/pkg/httpsrv"
	"context"
	"flag"
	"log"
	"os"
	"time"
)

var (
	action string
	host   = "0.0.0.0"
	port   = 8000
)

func init() {

	flag.StringVar(&action, "action", "seed", "seed pizza data on redis")
	flag.StringVar(&host, "host", host, "host address to serve web server")
	flag.IntVar(&port, "port", port, "port to expose our web server")
	flag.Parse()
}
func main() {

	ctx := context.Background()
	rdb := db.GetRedisDbClient(context.Background())

	switch action {
	case "web":
		srv := httpsrv.NewServer(host, port, nil)
		srv.Start()
	case "seed":
		start := time.Now()
		var path, _ = os.Getwd()
		model.Run(ctx, rdb, path, "data/pizzas.csv")
		elapsed := time.Since(start)
		log.Printf("Seed pizza data on redis [%s]\n", elapsed.String())
	}

}
