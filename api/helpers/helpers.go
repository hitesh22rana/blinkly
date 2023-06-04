package helpers

import (
	"os"
	"strings"
)

var Domain = os.Getenv("DOMAIN")

func EnforceHTTP(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "http://" + url
	}

	return url
}

func RemoveDomainError(url string) bool {
	if url == Domain {
		return false
	}

	newURL := strings.Replace(url, "http://", "", 1)
	newURL = strings.Replace(newURL, "https://", "", 1)
	newURL = strings.Replace(newURL, "www.", "", 1)
	newURL = strings.Split(newURL, "/")[0]

	return newURL != Domain
}
