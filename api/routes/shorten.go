package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/suraj/url-shortener/database"
	"github.com/suraj/url-shortener/helpers"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	Status          string        `json:"status"`
	Message         string        `json:"message"`
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	var body request

	// Decode the Request body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response{
			Status:          "false",
			Message:         "cannot parse JSON",
			URL:             "",
			CustomShort:     "",
			Expiry:          0,
			XRateRemaining:  0,
			XRateLimitReset: 0,
		})
		return
	}

	// Create a db client for Rate Limiting with number
	r2 := database.CreateClient(1)
	defer r2.Close()

	clientIP := r.RemoteAddr
	val, err := r2.Get(database.Ctx, clientIP).Result()

	if err == redis.Nil {
		_ = r2.Set(database.Ctx, clientIP, os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, clientIP).Result()
			limitInMinutes := int(limit / time.Nanosecond / time.Minute)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(response{
				Status:          "false",
				Message:         "Rate Limit Exceeded. Please try again after " + strconv.Itoa(limitInMinutes) + " minutes",
				URL:             "",
				CustomShort:     "",
				Expiry:          0,
				XRateRemaining:  0,
				XRateLimitReset: 0,
			})
			return
		}
	}

	// URL validation
	if govalidator.IsURL(body.URL) == false {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response{
			Status:          "false",
			Message:         "Invalid URL",
			URL:             "",
			CustomShort:     "",
			Expiry:          0,
			XRateRemaining:  0,
			XRateLimitReset: 0,
		})
		return
	}

	// Checks if the URL has same domain as shotrening service
	if helpers.IsSameDomain(body.URL) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response{
			Status:          "false",
			Message:         "Service unavailable for this URL",
			URL:             "",
			CustomShort:     "",
			Expiry:          0,
			XRateRemaining:  0,
			XRateLimitReset: 0,
		})
		return
	}

	body.URL = helpers.EnforceHTTP(body.URL) // adding prefix https:// to the url

	// generate short id or url
	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	// create a new db client to store url
	rdb := database.CreateClient(0)
	defer rdb.Close()

	for i := 0; i < 100; i++ {
		_, err = rdb.Get(database.Ctx, id).Result()
		if err == redis.Nil { // err found in searching means, so unique id found
			break
		} else if err != nil {
			// some other Redis error
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response{
				Status:          "false",
				Message:         "Internal Server Error",
				URL:             "",
				CustomShort:     "",
				Expiry:          0,
				XRateRemaining:  0,
				XRateLimitReset: 0,
			})
			return
		}

		// no error means, that id was already present
		id = uuid.New().String()[:6]

		if i == 99 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response{
				Status:          "false",
				Message:         "Could not generate unique URL",
				URL:             "",
				CustomShort:     "",
				Expiry:          0,
				XRateRemaining:  0,
				XRateLimitReset: 0,
			})
			return
		}

	}

	if body.Expiry == 0 { // In frontend 0 means empty, so set default expiry to 24 at backend
		body.Expiry = 24
	} else if body.Expiry == -1 { // In frontend -1 means never expire, so set 0 for never expire
		body.Expiry = 0 // In Redis, 0 means never expire
	}

	err = rdb.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response{
			Status:          "false",
			Message:         "Unable to connct to Server",
			URL:             "",
			CustomShort:     "",
			Expiry:          0,
			XRateRemaining:  0,
			XRateLimitReset: 0,
		})
		return
	}

	// Prepare Response
	resp := response{
		Status:          "true",
		Message:         "Url Shortened",
		URL:             body.URL,
		CustomShort:     os.Getenv("DOMAIN") + "/" + id,
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	// Reduce quota by 1
	r2.Decr(database.Ctx, clientIP)

	// update XRateRemaining & XRateLimitReset in response
	val, _ = r2.Get(database.Ctx, clientIP).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(database.Ctx, clientIP).Result()          // TTL : Total time limit
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute // convert time to minutes

	// Send Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
