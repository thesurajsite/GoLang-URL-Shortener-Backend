package database

// `package database` → Defines the package name as `database`.
// In Go, code is organized into packages.
// Everything in this file belongs to the `database` package.

// ---------------- Imports ----------------
import (
	"context" // Provides context management (cancellation, deadlines, etc.)
	"os"      // Gives access to environment variables & system functions

	"github.com/go-redis/redis/v8" // Redis client library for Go (version 8)
)

// ---------------- Global Variables ----------------
var Ctx = context.Background()

// `var` → Declares a variable.
// `Ctx` → A global context variable we’ll use for Redis commands.
// `context.Background()` → Creates an empty base context (no deadline, no cancel).
// Context is used in Go to manage request lifetimes across API boundaries.

// ---------------- Function ----------------
func CreateClient(dbNo int) *redis.Client {
	// `func` → Defines a function.
	// `CreateClient` → Function name, exported (capital letter).
	// `(dbNo int)` → Takes an integer parameter `dbNo` (the Redis DB index).
	// `*redis.Client` → Returns a pointer to a Redis client.

	rdb := redis.NewClient(&redis.Options{
		// Creates a new Redis client using options.

		Addr: os.Getenv("DB_ADDR"),
		// `Addr` → Redis server address (host:port).
		// `os.Getenv("DB_ADDR")` → Fetches `DB_ADDR` value from environment variables.

		Password: os.Getenv("DB_PASS"),
		// `Password` → Redis server password.
		// `os.Getenv("DB_PASS")` → Fetches `DB_PASS` from environment variables.

		DB: dbNo,
		// `DB` → The Redis database number (integer).
		// Redis allows multiple logical DBs inside one instance.
		// `dbNo` is passed as an argument when calling this function.
	})

	return rdb
	// Returns the Redis client object so it can be used in other parts of the code.
}
