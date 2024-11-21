package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/time/rate"
)

type PageDiscoveryScraper struct {
	client      *http.Client
	rateLimiter *rate.Limiter
	visited     sync.Map
	results     []string
	mu          sync.Mutex
}

func NewPageDiscoveryScraper() *PageDiscoveryScraper {
	return &PageDiscoveryScraper{
		client:      &http.Client{Timeout: 10 * time.Second},
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), 5),
		results:     []string{},
	}
}

func (s *PageDiscoveryScraper) DiscoverPages(baseURL string, maxDepth int) ([]string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	// Create a WaitGroup to manage concurrent crawling
	var wg sync.WaitGroup
	wg.Add(1)

	// Start discovery from base URL
	go s.crawlPage(parsedURL, 0, maxDepth, &wg)

	// Wait for all crawling to complete
	wg.Wait()

	// Convert results to slice
	uniquePages := s.getUniquePages()
	return uniquePages, nil
}

func (s *PageDiscoveryScraper) crawlPage(pageURL *url.URL, currentDepth, maxDepth int, wg *sync.WaitGroup) {
	defer wg.Done()

	// Stop if max depth reached
	if currentDepth > maxDepth {
		return
	}

	// Check if already visited
	if _, visited := s.visited.LoadOrStore(pageURL.String(), true); visited {
		return
	}

	// Rate limit
	err := s.rateLimiter.Wait(context.Background())
	if err != nil {
		fmt.Printf("Rate limit error: %v\n", err)
		return
	}

	// Fetch page
	resp, err := s.client.Get(pageURL.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Only process HTML pages
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return
	}

	// Parse document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	// Store current page
	s.mu.Lock()
	s.results = append(s.results, pageURL.String())
	s.mu.Unlock()

	// Find all links
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		link, exists := sel.Attr("href")
		if !exists {
			return
		}

		// Normalize URL
		linkURL, err := url.Parse(link)
		if err != nil {
			return
		}

		// Make absolute URL
		absoluteURL := pageURL.ResolveReference(linkURL)

		// Filter to same domain and http/https
		if absoluteURL.Hostname() != pageURL.Hostname() ||
			!(absoluteURL.Scheme == "http" || absoluteURL.Scheme == "https") {
			return
		}

		// Skip fragments and query parameters
		absoluteURL.Fragment = ""
		absoluteURL.RawQuery = ""

		// Recursive crawl
		wg.Add(1)
		go s.crawlPage(absoluteURL, currentDepth+1, maxDepth, wg)
	})
}

func (s *PageDiscoveryScraper) getUniquePages() []string {
	uniqueMap := make(map[string]bool)
	var uniquePages []string

	s.visited.Range(func(key, value interface{}) bool {
		pageURL := key.(string)
		if !uniqueMap[pageURL] {
			uniqueMap[pageURL] = true
			uniquePages = append(uniquePages, pageURL)
		}
		return true
	})

	return uniquePages
}
