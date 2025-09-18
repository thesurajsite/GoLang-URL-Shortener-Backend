package routes

// Package name → `routes`
// This package contains HTTP route handlers for the URL shortener service.

import (
	"net/http" // Provides HTTP status codes and server utilities

	"github.com/gin-gonic/gin"                // Gin web framework for building HTTP APIs
	"github.com/go-redis/redis/v8"            // Redis client library for Go (v8)
	"github.com/suraj/url-shortener/database" // Custom package to handle Redis connections
)

// ---------------- Route Handler ----------------
func ResolveURL(c *gin.Context) {
	// `func` → defines a function
	// `ResolveURL` → route handler for resolving short URLs
	// `(c *gin.Context)` → parameter: Gin context for request/response

	url := c.Param("url")
	// `c.Param("url")` → extracts the URL parameter from the request path
	// Example: if the path is "/abc123", url = "abc123"

	// ---------------- Connect to Redis DB 0 ----------------
	r := database.CreateClient(0)
	// Calls the `CreateClient` function from database package
	// Connects to Redis database 0 (used for storing short->long URL mappings)

	defer r.Close()
	// Ensures Redis connection is closed when function exits

	// ---------------- Fetch the long URL ----------------
	value, err := r.Get(database.Ctx, url).Result()
	// `r.Get` → fetches the value associated with the key `url` from Redis
	// `database.Ctx` → context passed to Redis command
	// `.Result()` → returns the actual value or error

	if err == redis.Nil {
		// `redis.Nil` → means the key does not exist in Redis
		c.JSON(http.StatusNotFound, gin.H{"error": "short not found in the database"})
		// Returns 404 status with JSON error message
		return
	} else if err != nil {
		// Any other Redis error (e.g., connection failure)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot connect to DB"})
		// Returns 500 status with JSON error message
		return
	}

	// ---------------- Increment counter in Redis DB 1 ----------------
	rInr := database.CreateClient(1)
	// Connect to Redis database 1 (used for analytics, e.g., counter)

	defer rInr.Close()
	// Close this Redis connection when function exits

	_ = rInr.Incr(database.Ctx, "counter")
	// Increment a counter key in Redis to track total redirects
	// `_ =` → ignoring the result/error

	// ---------------- Redirect to the original URL ----------------
	c.Redirect(http.StatusMovedPermanently, value)
	// `c.Redirect` → sends HTTP 301 redirect
	// `http.StatusMovedPermanently` → 301 status code
	// `value` → the long/original URL retrieved from Redis
}
