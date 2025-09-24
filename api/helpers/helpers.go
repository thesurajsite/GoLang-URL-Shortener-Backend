package helpers

import (
	"os"
	"strings"
)

// Add https:// to the url
func EnforceHTTP(url string) string {
	if !strings.HasPrefix(url, "http") {
		return "http://" + url
	}
	return url
}

// Check if URL Domain same as shortener
func IsSameDomain(url string) bool {

	if url == os.Getenv("DOMAIN") {
		return false
	}

	newURL := strings.TrimPrefix(url, "http://")
	newURL = strings.TrimPrefix(newURL, "https://")
	newURL = strings.TrimPrefix(newURL, "www.")
	newURL = strings.Split(newURL, "/")[0]
	return newURL == os.Getenv("DOMAIN")

}
