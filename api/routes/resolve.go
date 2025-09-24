package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/suraj/url-shortener/database"
)

// Response struct for ResolveURL
type ResolveResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

func ResolveURL(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	url := params["id"] // Read {id} from the URL

	rd := database.CreateClient(0) // Connects to db that stores url
	defer rd.Close()

	value, err := rd.Get(database.Ctx, url).Result()

	if err == redis.Nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ResolveResponse{
			Status:  "false",
			Message: "Short URL not found in the database",
			URL:     "",
		})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ResolveResponse{
			Status:  "false",
			Message: "Cannot connect to server",
			URL:     "",
		})
		return
	}

	rInr := database.CreateClient(1)
	defer rInr.Close()

	_ = rInr.Incr(database.Ctx, "counter")

	// Redirects user to New URL
	http.Redirect(w, r, value, http.StatusMovedPermanently)
}
