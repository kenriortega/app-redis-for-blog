package main

import (
	"app/pkg/db"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

const (
	reqLimit            = 10
	durationLimit       = 20
	ipLimiterName       = "ip-limiter"
	keyIPRequestsPrefix = "requests"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "Port to serve")
	flag.Parse()
	rdb := db.GetRedisDbClient(context.TODO())
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		ip := extractIpAddr(r)
		controller := NewIPController(rdb)
		requests, accepted := controller.AcceptedRequest(
			context.TODO(),
			ip,
			reqLimit,
			durationLimit,
		)
		if !accepted {
			w.WriteHeader(http.StatusTooManyRequests)
		}
		w.Header().Add("X-RateLimit-Limit", strconv.Itoa(reqLimit))
		w.Header().Add("X-RateLimit-Remaining", strconv.Itoa(10-requests))

	})

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

type IPController struct {
	rdb *redis.Client
}

func NewIPController(rdb *redis.Client) *IPController {
	return &IPController{
		rdb: rdb,
	}
}
func (c *IPController) createKEY(ip string) string {
	return fmt.Sprintf("%s:%s", keyIPRequestsPrefix, ip)
}

func (c *IPController) AcceptedRequest(ctx context.Context, ip string, limit, limitDuration int) (int, bool) {
	key := c.createKEY(ip)

	if _, err := c.rdb.Get(ctx, key).Result(); err == redis.Nil {
		err := c.rdb.Set(ctx, key, "0", time.Second*time.Duration(limitDuration))
		if err != nil {
			log.Println(err)
			return 0, false
		}
	}

	if _, err := c.rdb.Incr(ctx, key).Result(); err != nil {
		log.Println(err)
		return 0, false
	}

	requests, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		log.Println(err)
		return 0, false
	}
	requestsNum, err := strconv.Atoi(requests)
	if err != nil {
		log.Println(err)
		return 0, false
	}

	if requestsNum > limit {
		return requestsNum, false
	}

	return requestsNum, true

}

func extractIpAddr(req *http.Request) string {
	ipAddress := req.RemoteAddr
	fwdAddress := req.Header.Get("X-Forwarded-For") // capitalisation doesn't matter
	if fwdAddress != "" {
		// Got X-Forwarded-For
		ipAddress = fwdAddress // If it's a single IP, then awesome!

		// If we got an array... grab the first IP
		ips := strings.Split(fwdAddress, ", ")
		if len(ips) > 1 {
			ipAddress = ips[0]
		}
	}
	remoteAddrToParse := ""
	if strings.Contains(ipAddress, "[::1]") {
		remoteAddrToParse = strings.Replace(ipAddress, "[::1]", "localhost", -1)
		ipAddress = strings.Split(remoteAddrToParse, ":")[0]
	} else {
		ipAddress = strings.Split(ipAddress, ":")[0]
	}
	return ipAddress
}
