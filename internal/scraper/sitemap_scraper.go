package scraper

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

type SitemapScraper struct {
	client      *http.Client
	rateLimiter *rate.Limiter
}

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod,omitempty"`
	ChangeFreq string  `xml:"changefreq,omitempty"`
	Priority   float64 `xml:"priority,omitempty"`
}

func NewSitemapScraper() *SitemapScraper {
	return &SitemapScraper{
		client:      &http.Client{Timeout: 30 * time.Second},
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), 2),
	}
}

func (s *SitemapScraper) ScrapeSitemap(config ScraperConfig) ([]SitemapResult, error) {
	// Normalize sitemap URL
	sitemapURL := config.URL
	if !strings.HasSuffix(sitemapURL, "sitemap.xml") {
		sitemapURL = strings.TrimSuffix(sitemapURL, "/") + "/sitemap.xml"
	}

	// Wait for rate limit
	err := s.rateLimiter.Wait(context.Background())
	if err != nil {
		return nil, fmt.Errorf("rate limit error: %v", err)
	}

	// Fetch sitemap
	resp, err := s.client.Get(sitemapURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sitemap: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse sitemap
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sitemap body: %v", err)
	}

	var urlSet URLSet
	err = xml.Unmarshal(body, &urlSet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sitemap XML: %v", err)
	}

	// Convert to SitemapResult
	results := make([]SitemapResult, len(urlSet.URLs))
	for i, u := range urlSet.URLs {
		results[i] = SitemapResult{
			URL:          u.Loc,
			LastModified: u.LastMod,
			ChangeFreq:   u.ChangeFreq,
			Priority:     u.Priority,
		}
	}

	return results, nil
}
