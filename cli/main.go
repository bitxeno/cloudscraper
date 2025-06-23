package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Advik-B/cloudscraper/lib"
)

func main() {
	fmt.Println("Attempting to create a scraper that uses external Node.js...")

	// Create a new scraper instance, specifically configuring it to use "node".
	// The library will automatically find the 'node' executable in your system's PATH.
	sc, err := cloudscraper.New(
		cloudscraper.WithJSRuntime("node"),
	)
	if err != nil {
		// This error will trigger if 'node' is not found in the PATH.
		log.Fatalf("Failed to create scraper: %v. Is Node.js installed and in your PATH?", err)
		os.Exit(1)
	}

	fmt.Println("Scraper created successfully. Making request to a Cloudflare-protected site...")

	// A site known to be protected by Cloudflare's JS challenge
	targetURL := "https://nowsecure.nl"

	// Make a GET request
	resp, err := sc.Get(targetURL)
	if err != nil {
		log.Fatalf("Request to %s failed: %v", targetURL, err)
	}
	defer resp.Body.Close()

	fmt.Printf("\n--- Response ---\n")
	fmt.Printf("Status: %s\n", resp.Status)

	// Read and print the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// Print a preview of the HTML to confirm success
	preview := string(body)
	if len(preview) > 500 {
		preview = preview[:500]
	}
	fmt.Printf("Body Preview:\n%s...\n", preview)
	fmt.Println("----------------")

	if resp.StatusCode == 200 {
		fmt.Println("\nSuccess! Cloudflare challenge was bypassed using Node.js.")
	} else {
		fmt.Printf("\nFailed to bypass challenge. Received status code: %d\n", resp.StatusCode)
	}
}
