package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"bgsearcher.com/util"
)

// Ez2Play is a struct for shop
type Ez2Play struct {
	Info     ShopInfo
	CacheMap map[string]SearchCache
}

// GetShopInfo returns the shop's info
func (s Ez2Play) GetShopInfo() ShopInfo {
	return s.Info
}

// UpdatePrevNewArrivals for specific shops
func (s Ez2Play) UpdatePrevNewArrivals(arrivals []NewArrival) {
	return
}

// GetSearchResults is an exported method of Crawler by Ez2Play
func (s Ez2Play) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	if val, exists := s.CacheMap[query]; exists {
		now := time.Now()
		if now.Sub(val.SearchedTime) <= searchCacheDuration {
			return val.Results
		}
	}

	req, err := http.NewRequest("GET", info.QueryURL+url.QueryEscape(util.ToEUCKR(query)), nil)
	if err != nil {
		log.Printf("Ez2Play: Failed to make request: %s", err)
		return results
	}

	client := &http.Client{
		Timeout: timeoutDuration,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ez2Play: Failed to get page: %s", err)
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Ez2Play: Failed to read response")
		return results
	}

	doc.Find("tbody").Each(func(i int, s *goquery.Selection) {
		// each item
		aTag := s.Children().Children().Eq(0).Children().Children().Children().Children().Children()
		url, exists := aTag.Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + url
		if !strings.Contains(url, "product") {
			return
		}

		img, exists := aTag.Children().Attr("src")
		if !exists {
			return
		}
		if !strings.Contains(img, "tiny") {
			return
		}

		name1 := s.Children().Children().Eq(1).Find("a").Eq(0).Find("font").Text()
		name1 = util.ToUTF8(name1)

		name2 := ""

		price := s.Children().Children().Eq(3).Find("font").Eq(0).Text()
		if strings.Contains(price, "₩") {
			price = strings.TrimSpace(price)
			price = strings.Split(price, "₩")[1]
			price = util.ToUTF8(price)
			price += "원"
		}

		var soldOut = false
		s.Children().Children().Eq(1).Find("img").Each(func(j int, ss *goquery.Selection) {
			if src, exists := ss.Attr("src"); exists {
				if strings.Contains(src, "sellout") {
					soldOut = true
				}
			}
		})

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	s.CacheMap[query] = SearchCache{time.Now(), results}

	return results
}

// GetNewArrivals is an exported method of Crawler by Ez2Play
func (s Ez2Play) GetNewArrivals() []NewArrival {
	var results []NewArrival
	return results
}
