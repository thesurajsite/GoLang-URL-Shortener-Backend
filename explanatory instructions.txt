URL-SHORTENER/
│── .data/                    # runtime Redis data (ignored in git)
│
├── api/
│   ├── database/
│   │   └── database.go       # Redis connection logic
│   │
│   ├── helpers/
│   │   └── helpers.go        # Utility functions
│   │
│   └── routes/
│       ├── resolve.go        # handler: resolve short → long
│       ├── shorten.go        # handler: shorten URL
│       └── routes.go         # Setup() function to register routes
│
│── .env                      # Environment variables
│── Dockerfile                # API Dockerfile
│── docker-compose.yml         # Orchestrates API + Redis
│── go.mod
│── go.sum
│── main.go                   # Entry point, loads env + starts Gin server
│
└── db/
    └── Dockerfile            # Optional, if you want custom Redis build










FLOW

1. Starting the server

 `main.go` is the entry point.
 Loads environment variables from `.env` (like `APP_PORT`, Redis credentials, API quota, domain).
 Creates a Gin router (`gin.Default()`).
 Calls `routes.Setup(r)` to register your routes.
 Starts the server on the port from `APP_PORT`.

---

2. Shortening a URL (`ShortenURL`)

 The user sends a POST request with JSON containing: `url`, optional `short` name, optional `expiry`.
 The server parses this JSON into a `request` struct.
 Rate limiting:

   Redis DB 1 stores how many requests the user (IP) can still make.
   If quota is exceeded, return a “Rate Limit Exceeded” error.
 URL validation:

   Checks the URL format.
   Prevents shortening the service’s own domain.
   Ensures `http://` is added if missing.
 Short ID generation:

   Uses the user’s custom short ID if provided, or generates a random 6-character string.
 Check Redis DB 0:

   Ensures the short ID isn’t already used.
   Saves the mapping: `short ID → original URL` with expiry.
 Update rate limit:

   Decreases the user’s quota in Redis DB 1.
 Respond:

   Returns JSON with the short URL, expiry, and remaining quota info.

---

3. Resolving a short URL (`ResolveURL`)

 The user visits a short URL (`example.com/abc123`).
 The server extracts the short ID from the path.
 Fetch original URL:

   Redis DB 0 stores short ID → original URL.
   If not found → return 404 error.
 Update redirect counter:

   Redis DB 1 increments a counter for analytics.
 Redirect:

   The user is redirected to the original URL with HTTP 301 status.

---

4. Helper functions

 `EnforceHTTP(url)` → ensures the URL starts with `http://`.
 `RemoveDomainError(url)` → prevents shortening URLs that point to your own domain.

---

5. Database connections

 `database.CreateClient(dbNo)` creates a Redis client.
 DB 0 → stores URL mappings.
 DB 1 → stores API rate limits and counters.
 A global `Ctx` is used for Redis operations.

---

In short:

 POST /shorten → create short URL, save in Redis, track quota.
 GET /\:shortID → fetch original URL from Redis, increment counter, redirect.
