package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator" // URL validation
	"github.com/go-redis/redis/v8"      // Redis client
	"github.com/google/uuid"            // UUID for unique short IDs
	"github.com/suraj/url-shortener/database"
	"github.com/suraj/url-shortener/helpers"
)

// ---------------- Request & Response Models ----------------

// Incoming request body structure
type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

// Outgoing response structure
type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

// ---------------- Route Handler ----------------

// ShortenURL handles POST /api/v1/ requests for shortening URLs
func ShortenURL(w http.ResponseWriter, r *http.Request) {
	var body request

	// Decode incoming JSON into request struct
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"cannot parse JSON"}`, http.StatusBadRequest)
		return
	}

	// ---------------- Rate Limiting ----------------
	r2 := database.CreateClient(1) // Redis DB 1 for quota
	defer r2.Close()

	clientIP := r.RemoteAddr
	val, err := r2.Get(database.Ctx, clientIP).Result()

	if err == redis.Nil {
		// First request from this IP → set quota
		_ = r2.Set(database.Ctx, clientIP, os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, clientIP).Result()
			resp := map[string]interface{}{
				"error":            "Rate Limit Exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}

	// ---------------- URL Validation ----------------
	if !govalidator.IsURL(body.URL) {
		http.Error(w, `{"error":"Invalid URL"}`, http.StatusBadRequest)
		return
	}

	if !helpers.RemoveDomainError(body.URL) {
		http.Error(w, `{"error":"Service not available"}`, http.StatusServiceUnavailable)
		return
	}

	body.URL = helpers.EnforceHTTP(body.URL)

	// ---------------- Generate Short ID ----------------
	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	// ---------------- Store in Redis DB 0 ----------------
	rdb := database.CreateClient(0)
	defer rdb.Close()

	val, _ = rdb.Get(database.Ctx, id).Result()
	if val != "" {
		http.Error(w, `{"error":"URL custom short is already in use"}`, http.StatusForbidden)
		return
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = rdb.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		http.Error(w, `{"error":"Unable to connect to server"}`, http.StatusInternalServerError)
		return
	}

	// ---------------- Prepare Response ----------------
	resp := response{
		URL:             body.URL,
		CustomShort:     os.Getenv("DOMAIN") + "/" + id,
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	// Reduce quota by 1
	r2.Decr(database.Ctx, clientIP)

	val, _ = r2.Get(database.Ctx, clientIP).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(database.Ctx, clientIP).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	// ---------------- Send JSON Response ----------------
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
