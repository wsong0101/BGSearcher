package crawl

// SearchResult represents individual game info
type SearchResult struct {
	Company string
	Link    string
	Img     string
	Name    string
	Name2   string
	Price   string
	SoldOut bool
}

// ShopInfo represents each shop's information
type ShopInfo struct {
	URL          string
	Name         string
	QueryURL     string
	LinkPrefix   string
	FireStoreDir string
}

// Crawler is the interface for boardgame shops' funcs
type Crawler interface {
	GetSearchResults(query string) []SearchResult
}
