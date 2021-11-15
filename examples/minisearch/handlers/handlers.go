package handlers

import (
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

func writeResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
