package handlers

import (
	"app/examples/minisearch/domain"
	"encoding/json"
	"fmt"
	"net/http"

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
	result, err := h.rdb.Do(
		r.Context(),
		"FT.SEARCH",
		domain.INDEX,
		fmt.Sprintf(`@country:{%s}`, params["country"]),
		"LIMIT",
		0,
		100,
	).Result()
	total := result.([]interface{})[0]
	docs := result.([]interface{})[1:]
	var values []interface{}
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

	data := make(map[string]interface{})
	data["total"] = total
	data["total_peer_page"] = len(values)
	data["docs"] = values
	if err != nil {
		writeResponse(w, http.StatusBadRequest, err)
	} else {
		writeResponse(w, http.StatusOK, data)
	}
}

func writeResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
