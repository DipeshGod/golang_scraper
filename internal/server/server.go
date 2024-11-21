package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dipeshgod/go-scraper/internal/scraper"
)

type ScraperServer struct {
	GoogleScraper  *scraper.GoogleScraper
	SitemapScraper *scraper.SitemapScraper
}

func NewScraperServer() *ScraperServer {
	return &ScraperServer{
		GoogleScraper:  scraper.NewGoogleScraper(),
		SitemapScraper: scraper.NewSitemapScraper(),
	}
}

func (s *ScraperServer) HandleGoogleScrape(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config scraper.ScraperConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results, err := s.GoogleScraper.Scrape(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *ScraperServer) HandleSitemapScrape(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config scraper.ScraperConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results, err := s.SitemapScraper.ScrapeSitemap(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *ScraperServer) HandlePageDiscovery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type DiscoveryRequest struct {
		URL      string `json:"url"`
		MaxDepth int    `json:"max_depth"`
	}

	var req DiscoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Default max depth if not provided
	if req.MaxDepth == 0 {
		req.MaxDepth = 3
	}

	pageDiscoveryScraper := scraper.NewPageDiscoveryScraper()
	pages, err := pageDiscoveryScraper.DiscoverPages(req.URL, req.MaxDepth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pages)
}

func StartServer(addr string) error {
	scraperServer := NewScraperServer()

	http.HandleFunc("/scrape/google", scraperServer.HandleGoogleScrape)
	http.HandleFunc("/scrape/sitemap", scraperServer.HandleSitemapScrape)
	http.HandleFunc("/scrape/pages", scraperServer.HandlePageDiscovery)

	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, nil)
}
