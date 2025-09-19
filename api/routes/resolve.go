package routes

// Package name → `routes`
// This package contains HTTP route handlers for the URL shortener service.

import (
	"encoding/json"
	"net/http" // Provides HTTP request/response handling

	"github.com/go-redis/redis/v8"            // Redis client library
	"github.com/gorilla/mux"                  // Gorilla Mux router
	"github.com/suraj/url-shortener/database" // Custom package to handle Redis connections
)

// ---------------- Route Handler ----------------
func ResolveURL(w http.ResponseWriter, r *http.Request) {
	// `func` → defines a function
	// `ResolveURL` → route handler for resolving short URLs
	// `(w http.ResponseWriter, r *http.Request)` → standard Go HTTP handler signature

	params := mux.Vars(r)
	url := params["url"]
	// `mux.Vars(r)` → extracts route variables from the URL path
	// Example: if the path is "/abc123", then url = "abc123"

	// ---------------- Connect to Redis DB 0 ----------------
	rd := database.CreateClient(0)
	// Connects to Redis database 0 (used for storing short->long URL mappings)

	defer rd.Close()
	// Ensures Redis connection is closed when function exits

	// ---------------- Fetch the long URL ----------------
	value, err := rd.Get(database.Ctx, url).Result()
	// `rd.Get` → fetches the value associated with the key `url` from Redis
	// `.Result()` → returns the actual value or error

	if err == redis.Nil {
		// `redis.Nil` → means the key does not exist in Redis
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "short not found in the database"})
		return
	} else if err != nil {
		// Any other Redis error (e.g., connection failure)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Cannot connect to DB"})
		return
	}

	// ---------------- Increment counter in Redis DB 1 ----------------
	rInr := database.CreateClient(1)
	// Connect to Redis database 1 (used for analytics, e.g., counter)

	defer rInr.Close()

	_ = rInr.Incr(database.Ctx, "counter")
	// Increment a counter key in Redis to track total redirects

	// ---------------- Redirect to the original URL ----------------
	http.Redirect(w, r, value, http.StatusMovedPermanently)
	// `http.Redirect` → sends HTTP 301 redirect
	// `value` → the long/original URL retrieved from Redis
}
