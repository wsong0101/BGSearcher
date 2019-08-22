package crawl

import (
	"time"
)

var newArrivalsLimit float64 = 20 * 24 // hours
var timeoutDuration = time.Duration(5 * time.Second)
var searchCacheDuration = time.Duration(30 * time.Minute)

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

// SearchCache is struct to save searchresults
type SearchCache struct {
	SearchedTime time.Time
	Results      []SearchResult
}

// NewArrival represents arrived games at the time
type NewArrival struct {
	UpTime  time.Time
	Results []SearchResult
}

// ShopInfo represents each shop's information
type ShopInfo struct {
	URL           string
	Name          string
	QueryURL      string
	LinkPrefix    string
	FireStoreDir  string
	NewArrivalURL string
}

// Crawler is the interface for boardgame shops' funcs
type Crawler interface {
	GetSearchResults(query string) []SearchResult
	GetNewArrivals() []NewArrival
	GetShopInfo() ShopInfo
	UpdatePrevNewArrivals(arrivals []NewArrival)
}

func isEqualSearchResults(l []SearchResult, r []SearchResult) bool {
	if len(l) != len(r) {
		return false
	}

	for i := 0; i < len(l); i++ {
		var lElem = l[i]
		var rElem = r[i]
		if lElem.Name != rElem.Name {
			return false
		}
	}

	return true
}
