package routes

// Package `routes` contains HTTP route handlers for the URL shortener service.

import (
	"net/http" // Provides HTTP status codes and server utilities
	"os"       // Access environment variables
	"strconv"  // Convert strings to integers and vice versa
	"time"     // Time utilities for durations

	"github.com/asaskevich/govalidator"       // URL validation library
	"github.com/gin-gonic/gin"                // Gin web framework
	"github.com/go-redis/redis/v8"            // Redis client for Go
	"github.com/google/uuid"                  // Generate UUIDs
	"github.com/suraj/url-shortener/database" // Custom package for Redis connections
	"github.com/suraj/url-shortener/helpers"  // Helper functions for URL manipulation
)

// ---------------- Request & Response Structs ----------------
type request struct {
	URL         string        `json:"url"`    // URL to shorten
	CustomShort string        `json:"short"`  // Optional custom short URL
	Expiry      time.Duration `json:"expiry"` // Expiry time in hours
}

type response struct {
	URL             string        `json:"url"`              // Original URL
	CustomShort     string        `json:"short"`            // Short URL returned to user
	Expiry          time.Duration `json:"expiry"`           // Expiry in hours
	XRateRemaining  int           `json:"rate_limit"`       // Remaining API quota
	XRateLimitReset time.Duration `json:"rate_limit_reset"` // Time until quota resets
}

// ---------------- Route Handler ----------------
func ShortenURL(c *gin.Context) {
	var body request
	// `var body request` → create a variable to hold the incoming JSON request

	if err := c.ShouldBindJSON(&body); err != nil {
		// Bind incoming JSON to `body` struct
		// If JSON parsing fails:
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot parse JSON"})
		return
	}

	// ---------------- Rate Limiting ----------------
	r2 := database.CreateClient(1)
	// Connect to Redis DB 1 for API quota tracking
	defer r2.Close()

	val, err := r2.Get(database.Ctx, c.ClientIP()).Result()
	// Get the remaining quota for the client's IP address

	if err == redis.Nil {
		// If IP is new (no quota key exists)
		_ = r2.Set(database.Ctx, c.ClientIP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
		// Initialize quota with `API_QUOTA` from env for 30 minutes
	} else {
		valInt, _ := strconv.Atoi(val) // Convert string quota to integer
		if valInt <= 0 {
			// Quota exhausted
			limit, _ := r2.TTL(database.Ctx, c.ClientIP()).Result() // Get remaining TTL
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":            "Rate Limit Exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
			return
		}
	}

	// ---------------- URL Validation ----------------
	if !govalidator.IsURL(body.URL) {
		// Validate the URL format
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	if !helpers.RemoveDomainError(body.URL) {
		// Prevent shortening own domain URLs
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service not available"})
		return
	}

	body.URL = helpers.EnforceHTTP(body.URL)
	// Ensure URL starts with "http://"

	// ---------------- Generate Short ID ----------------
	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6] // Generate 6-char random ID
	} else {
		id = body.CustomShort // Use provided custom short
	}

	// ---------------- Store in Redis DB 0 ----------------
	r := database.CreateClient(0)
	defer r.Close()

	val, _ = r.Get(database.Ctx, id).Result()
	if val != "" {
		// Prevent overwriting existing short URL
		c.JSON(http.StatusForbidden, gin.H{"error": "URL custom short is already in use"})
		return
	}

	if body.Expiry == 0 {
		body.Expiry = 24 // Default expiry 24 hours
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to server"})
		return
	}

	// ---------------- Prepare Response ----------------
	resp := response{
		URL:             body.URL,
		CustomShort:     os.Getenv("DOMAIN") + "/" + id,
		Expiry:          body.Expiry,
		XRateRemaining:  10, // Default placeholder, will update below
		XRateLimitReset: 30, // Default placeholder, will update below
	}

	r2.Decr(database.Ctx, c.ClientIP()) // Reduce quota by 1

	val, _ = r2.Get(database.Ctx, c.ClientIP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val) // Update remaining quota

	ttl, _ := r2.TTL(database.Ctx, c.ClientIP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute // Update reset time

	// ---------------- Send JSON Response ----------------
	c.JSON(http.StatusOK, resp)
}
