package handlers

import (
	"app/examples/minisearch/domain"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// Handler ...
type Handler struct {
	rdb *redis.Client
}

// New ...
func New(rdb *redis.Client) Handler {
	return Handler{
		rdb: rdb,
	}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	info := make(map[string]interface{})
	info["version"] = "v0.0.1"
	info["name"] = "search-pizzas"

	writeResponse(w, http.StatusOK, info)
}

func (h *Handler) FindPizzaByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	result, err := h.rdb.HGetAll(
		r.Context(),
		fmt.Sprintf(`pizza:%s`, params["id"]),
	).Result()
	if err != nil {
		writeResponse(w, http.StatusBadRequest, err)
	} else {
		writeResponse(w, http.StatusOK, result)
	}
}

func (h *Handler) FindPizzasByCountry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	data, err := findQuery(
		r.Context(),
		h.rdb,
		fmt.Sprintf(`@country:{%s}`, params["country"]),
	)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, err)
	} else {
		writeResponse(w, http.StatusOK, data)
	}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query().Get("q")
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	var query string
	if start != "" && end != "" {

		dateStart, err := time.Parse("2006-01-02", start)
		if err != nil {
			log.Fatal(err)
		}
		dateEnd, err := time.Parse("2006-01-02", end)
		if err != nil {
			log.Fatal(err)
		}
		query = fmt.Sprintf(
			`%s @date_added:[%d %d]`, q, dateStart.Unix(), dateEnd.Unix(),
		)
	} else {
		query = q
	}
	data, err := findQuery(
		r.Context(),
		h.rdb,
		query,
	)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, err)
	} else {
		writeResponse(w, http.StatusOK, data)
	}
}

func (h *Handler) FindNearPizzas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	longitude := r.URL.Query().Get("lon")
	latitude := r.URL.Query().Get("lat")
	radius := r.URL.Query().Get("r")
	unit := r.URL.Query().Get("u")

	data, err := findQuery(
		r.Context(),
		h.rdb,
		fmt.Sprintf(
			`@location:[%s %s %s %s]`,
			longitude,
			latitude,
			radius,
			unit,
		),
	)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, err)
	} else {
		writeResponse(w, http.StatusOK, data)
	}
}

func (h *Handler) StatsByDate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	date := r.URL.Query().Get("date")
	dateStats, err := time.Parse("2006-01-02", date)
	if err != nil {
		log.Fatal(err)
	}

	data, err := aggregationQueryBytime(
		r.Context(),
		h.rdb,
		fmt.Sprintf(
			`@date_added:[%d +inf]`,
			dateStats.Unix(),
		),
		"@date_added",
	)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, err)
	} else {
		writeResponse(w, http.StatusOK, data)
	}
}

// utils functions

func findQuery(ctx context.Context, rdb *redis.Client, query string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var values []interface{}

	result, err := rdb.Do(
		ctx,
		"FT.SEARCH",
		domain.INDEX,
		query,
		"LIMIT",
		0,
		100,
	).Result()
	if err != nil {
		return nil, err
	}
	total := result.([]interface{})[0]
	docs := result.([]interface{})[1:]

	for i, doc := range docs {
		if i%2 != 0 {
			value := make(map[string]interface{})
			var k, v string
			for j, it := range doc.([]interface{}) {
				if j%2 == 0 {
					k = it.(string)
				}
				if j%2 != 0 {
					v = it.(string)
				}
				value[k] = v
			}
			values = append(values, value)
		}
	}

	data["total"] = total
	data["total_peer_page"] = len(values)
	data["docs"] = values
	return data, nil
}

func aggregationQueryBytime(
	ctx context.Context,
	rdb *redis.Client,
	query string,
	aggregateField string,
) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var values []interface{}

	result, err := rdb.Do(
		ctx,
		"FT.AGGREGATE",
		domain.INDEX,
		query,
		"GROUPBY",
		1,
		aggregateField,
		"REDUCE",
		"COUNT",
		0,
		"SORTBY",
		2,
		aggregateField,
		"ASC",
		"LIMIT",
		0,
		100,
	).Result()
	if err != nil {
		return nil, err
	}
	total := result.([]interface{})[0]
	docs := result.([]interface{})[1:]

	for i, doc := range docs {
		if i%2 != 0 {
			value := make(map[string]interface{})
			var k, v string
			for j, it := range doc.([]interface{}) {
				if j%2 == 0 {
					k = it.(string)
				}
				if j%2 != 0 {
					v = it.(string)
				}
				value[k] = v
			}
			values = append(values, value)
		}
	}

	data["total"] = total
	data["total_peer_page"] = len(values)
	data["docs"] = values
	return data, nil
}
func writeResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
