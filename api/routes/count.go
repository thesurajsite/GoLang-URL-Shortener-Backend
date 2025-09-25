package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"

	"github.com/suraj/url-shortener/database"
)

// Returns total number of URL generated
func UrlCount(w http.ResponseWriter, r *http.Request) {

	rd := database.CreateClient(0)
	defer rd.Close()

	count, err := rd.Get(database.Ctx, "generated").Result()

	if err == redis.Nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "true",
			"message": "total count fetched",
			"count":   "0",
		})
		return

	} else if err != nil {
		log.Println("Redis error:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "false",
			"message": "Internal Server Error",
			"count":   "0",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "true",
		"message": "Total URLs generated",
		"count":   count,
	})
}
