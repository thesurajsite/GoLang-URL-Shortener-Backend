package helpers

// Package name → `helpers`
// This package contains utility/helper functions for working with URLs.

import (
	"os"      // Provides access to environment variables & OS functions
	"strings" // Provides string manipulation utilities
)

// ---------------- Function 1 ----------------
func EnforceHTTP(url string) string {
	// `func` → defines a function
	// `EnforceHTTP` → function name (exported since it starts with capital letter)
	// `(url string)` → parameter: takes a string named `url`
	// `string` after → return type: returns a string

	if !strings.HasPrefix(url, "http") {
		// `strings.HasPrefix(url, "http")`
		// → checks if the given string `url` starts with "http"
		// `!` → NOT operator, means condition is true if url does NOT start with "http"

		return "http://" + url
		// If no "http"/"https" prefix, prepend "http://" to the URL
	}
	return url
	// Otherwise, return the original URL as is
}

// ---------------- Function 2 ----------------
func RemoveDomainError(url string) bool {
	// `RemoveDomainError` → function name (exported)
	// `(url string)` → parameter: takes a string
	// `bool` after → return type: returns a boolean (true/false)

	if url == os.Getenv("DOMAIN") {
		// `os.Getenv("DOMAIN")` → fetches environment variable named `DOMAIN`
		// Checks if the passed `url` is exactly equal to the `DOMAIN`
		return false
		// If it’s the same domain → return false (no error)
	}

	// ---------------- Cleanup the URL ----------------
	newURL := strings.TrimPrefix(url, "http://")
	// Removes "http://" if present at start of string

	newURL = strings.TrimPrefix(newURL, "https://")
	// Removes "https://" if present

	newURL = strings.TrimPrefix(newURL, "www.")
	// Removes "www." if present

	newURL = strings.Split(newURL, "/")[0]
	// Splits string by "/" and takes the first part (domain only, no paths)

	// ---------------- Final Check ----------------
	return newURL != os.Getenv("DOMAIN")
	// Returns true if the domain is NOT equal to environment variable DOMAIN
	// Returns false if they are the same
}
