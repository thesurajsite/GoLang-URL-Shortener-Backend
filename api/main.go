package main

// `package main` → Defines the entry point of a Go application.
// The `main` package is special in Go; it tells the compiler this is an executable program.

import (
	"log" // Standard logging library
	"os"  // Access environment variables

	"github.com/gin-gonic/gin"              // Gin web framework for building HTTP APIs
	"github.com/joho/godotenv"              // Load environment variables from a `.env` file
	"github.com/suraj/url-shortener/routes" // Custom package containing route handlers
)

func main() {
	_ = godotenv.Load()
	// Load environment variables from `.env` file
	// `_ =` → ignoring any error returned by Load()
	// Example: APP_PORT, DB_ADDR, DB_PASS, DOMAIN, API_QUOTA

	r := gin.Default()
	// Create a new Gin router with default middleware (logger and recovery)
	// `r` → variable holding the router instance

	routes.Setup(r)
	// Call the `Setup` function from the `routes` package
	// This function should register all routes (like `/api/v1/` and `/resolve/:url`) with Gin

	log.Fatal(r.Run(":" + os.Getenv("APP_PORT")))
	// Start the Gin server on the port specified in the environment variable `APP_PORT`
	// `log.Fatal` → logs any error and exits if server fails to start
}
