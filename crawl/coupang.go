package crawl

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Coupang is a struct for shop
type Coupang struct {
	Info     ShopInfo
	CacheMap map[string]SearchCache
}

// GetShopInfo returns the shop's info
func (s Coupang) GetShopInfo() ShopInfo {
	return s.Info
}

// UpdatePrevNewArrivals for specific shops
func (s Coupang) UpdatePrevNewArrivals(arrivals []NewArrival) {
	return
}

// GetSearchResults is an exported method of Crawler
func (s Coupang) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	if val, exists := s.CacheMap[query]; exists {
		now := time.Now()
		if now.Sub(val.SearchedTime) <= searchCacheDuration {
			return val.Results
		}
	}

	req, err := http.NewRequest("GET", info.QueryURL+url.QueryEscape(query), nil)
	if err != nil {
		log.Printf("Coupang: Failed to make request: %s", err)
		return results
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux i686; rv:64.0) Gecko/20100101 Firefox/64.0")

	client := &http.Client{
		Timeout: timeoutDuration,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Coupang: Failed to get page: %s", err)
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Coupang: Failed to read response")
		return results
	}

	doc.Find(".search-product").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find("a").Eq(0).Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + url

		img, exists := s.Find(".search-product-wrap-img").Attr("src")
		if !exists {
			return
		}
		img = "https:" + img

		var soldOut = false
		if s.Find(".out-of-stock").Text() != "" {
			soldOut = true
		}

		name1 := s.Find(".name").Text()
		name2 := ""

		price := s.Find(".price-value").Text()

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	s.CacheMap[query] = SearchCache{time.Now(), results}

	return results
}

// GetNewArrivals is an exported method of Crawler by Weefun
func (s Coupang) GetNewArrivals() []NewArrival {
	var results []NewArrival
	return results
}
