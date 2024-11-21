package scraper

type ScraperConfig struct {
	Keywords    []string `json:"keywords"`
	NumResults  int      `json:"num_results"`
	Language    string   `json:"language"`
	CountryCode string   `json:"country_code"`
	URL         string   `json:"url"`
}

type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type SitemapResult struct {
	URL          string  `json:"url"`
	LastModified string  `json:"last_modified,omitempty"`
	ChangeFreq   string  `json:"change_freq,omitempty"`
	Priority     float64 `json:"priority,omitempty"`
}
