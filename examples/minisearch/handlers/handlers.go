package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis/v8"
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

func writeResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
