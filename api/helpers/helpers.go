package helpers

import (
	"os"
	"strings"
)

func EnforceHTTP(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "http://" + url
	}

	return url
}

func RemoveDomainError(url string) bool {
	DOMAIN := os.Getenv("DOMAIN")

	if url == DOMAIN {
		return false
	}

	newURL := strings.Replace(url, "http://", "", 1)
	newURL = strings.Replace(newURL, "https://", "", 1)
	newURL = strings.Replace(newURL, "www.", "", 1)
	newURL = strings.Split(newURL, "/")[0]

	return newURL != DOMAIN
}
