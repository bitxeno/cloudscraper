package cloudscraper

import (
	"fmt"
	"io"
	"log"
)

func main() {
	// Create a new scraper with default settings
	sc, err := New()
	if err != nil {
		log.Fatalf("Failed to create scraper: %v", err)
	}

	// Make a GET request
	resp, err := sc.Get("https://nowsecure.nl") // A site known to be protected by Cloudflare
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)

	// Read and print the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// Print a preview of the response
	preview := string(body)
	if len(preview) > 500 {
		preview = preview[:500]
	}
	fmt.Printf("Body Preview:\n%s...\n", preview)
}
