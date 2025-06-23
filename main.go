package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"go-cloudscraper/scraper"
	"go-cloudscraper/scraper/proxy"
	"go-cloudscraper/scraper/stealth"
	useragent "go-cloudscraper/scraper/user_agent"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("A standalone, feature-complete Go port of Cloudscraper.")
		fmt.Println("\nUsage: go run . <url>")
		fmt.Println("Example: go run . https://nowsecure.nl")
		os.Exit(1)
	}
	targetURL := os.Args[1]

	var scraperOptions []scraper.ScraperOption

	// Example of advanced configuration
	scraperOptions = append(scraperOptions,
		scraper.WithBrowser(useragent.Config{Browser: "chrome"}),
		scraper.WithStealth(stealth.Options{
			Enabled:         true,
			MinDelay:        2 * time.Second,
			MaxDelay:        5 * time.Second,
			HumanLikeDelays: true,
		}),
	)
	
	// Add proxies if you have them in environment variables
	var proxies []string
	for i := 1; ; i++ {
		p := os.Getenv(fmt.Sprintf("PROXY%d", i))
		if p == "" { break }
		proxies = append(proxies, p)
	}
	if len(proxies) > 0 {
		scraperOptions = append(scraperOptions, scraper.WithProxies(proxies, proxy.Random, 5*time.Minute))
	}

	sc, err := scraper.New(scraperOptions...)
	if err != nil {
		log.Fatalf("Failed to create scraper: %v", err)
	}

	fmt.Printf("Scraping %s...\n", targetURL)

	resp, err := sc.Get(targetURL)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("\n--- Response ---\n")
	fmt.Printf("Status: %s\n", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil { log.Fatalf("Failed to read response body: %v", err) }

	bodyPreview := string(body)
	if len(bodyPreview) > 500 {
		bodyPreview = bodyPreview[:500] + "..."
	}
	fmt.Printf("\nBody Preview:\n%s\n", bodyPreview)
}