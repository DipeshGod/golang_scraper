package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

type GoogleScraper struct {
	client      *http.Client
	rateLimiter *rate.Limiter
}

func NewGoogleScraper() *GoogleScraper {
	return &GoogleScraper{
		client:      &http.Client{Timeout: 30 * time.Second},
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), 5),
	}
}

func (gs *GoogleScraper) Scrape(config ScraperConfig) (map[string][]SearchResult, error) {
	results := make(map[string][]SearchResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, keyword := range config.Keywords {
		wg.Add(1)
		go func(kw string) {
			defer wg.Done()

			// Wait for rate limit
			err := gs.rateLimiter.Wait(context.Background())
			if err != nil {
				fmt.Printf("Rate limit error: %v\n", err)
				return
			}

			pageResults, err := gs.scrapeGooglePage(kw, config)
			if err != nil {
				fmt.Printf("Scraping error for keyword %s: %v\n", kw, err)
				return
			}

			mu.Lock()
			results[kw] = pageResults
			mu.Unlock()
		}(keyword)
	}

	wg.Wait()
	return results, nil
}

func (gs *GoogleScraper) scrapeGooglePage(keyword string, config ScraperConfig) ([]SearchResult, error) {
	url := fmt.Sprintf("https://www.google.com/search?q=%s&num=%d&hl=%s&gl=%s",
		url.QueryEscape(keyword), config.NumResults, config.Language, config.CountryCode)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := gs.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	doc.Find(".g").Each(func(i int, s *goquery.Selection) {
		title := s.Find("h3").Text()
		link, _ := s.Find("a").First().Attr("href")
		description := s.Find(".VwiC3b").Text()

		if title != "" && link != "" {
			results = append(results, SearchResult{
				Title:       title,
				URL:         link,
				Description: description,
			})
		}
	})

	return results, nil
}
