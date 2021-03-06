package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bgsearcher.com/cloud"
	"github.com/PuerkitoBio/goquery"
)

// BoardgameMall is a struct for shop
type BoardgameMall struct {
	Info     ShopInfo
	CacheMap map[string]SearchCache
}

// GetShopInfo returns the shop's info
func (s BoardgameMall) GetShopInfo() ShopInfo {
	return s.Info
}

// UpdatePrevNewArrivals for specific shops
func (s BoardgameMall) UpdatePrevNewArrivals(arrivals []NewArrival) {
	return
}

// GetSearchResults is an exported method of Crawler
func (s BoardgameMall) GetSearchResults(query string) []SearchResult {
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
		log.Printf("BoardgameMall: Failed to make request: %s", err)
		return results
	}

	client := &http.Client{
		Timeout: timeoutDuration,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("BoardgameMall: Failed to get page: %s", err)
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("BoardgameMall: Failed to read response")
		return results
	}

	doc.Find(".goods_list_cont").Find(".item_cont").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find(".item_photo_box").Find("a").Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + strings.Split(url, "..")[1]

		img, exists := s.Find(".item_photo_box").Find("img").Attr("src")
		if !exists {
			return
		}
		img = cloud.GetURLFromCloud(info.FireStoreDir+img, info.LinkPrefix+img)

		var soldOut = false

		name1 := s.Find(".item_info_cont").Find(".item_name").Text()

		name2 := ""

		price := s.Find(".item_money_box").Find(".item_price").Text()
		price = strings.TrimSpace(price)

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	s.CacheMap[query] = SearchCache{time.Now(), results}

	return results
}

// GetNewArrivals is an exported method of Crawler by BoardgameMall
func (s BoardgameMall) GetNewArrivals() []NewArrival {
	var results []NewArrival

	return results
}
