package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bgsearcher.com/cloud"
	"bgsearcher.com/util"
	"github.com/PuerkitoBio/goquery"
)

// HobbyGameMall is a struct for shop
type HobbyGameMall struct {
	Info     ShopInfo
	CacheMap map[string]SearchCache
}

// GetShopInfo returns the shop's info
func (s HobbyGameMall) GetShopInfo() ShopInfo {
	return s.Info
}

// UpdatePrevNewArrivals for specific shops
func (s HobbyGameMall) UpdatePrevNewArrivals(arrivals []NewArrival) {
	return
}

// GetSearchResults is an exported method of Crawler
func (s HobbyGameMall) GetSearchResults(query string) []SearchResult {
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
		log.Printf("HobbyGameMall: Failed to make request: %s", err)
		return results
	}

	client := &http.Client{
		Timeout: timeoutDuration,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HobbyGameMall: Failed to get page: %s", err)
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("HobbyGameMall: Failed to read response")
		return results
	}

	doc.Find(".indiv").Find("table").Find("table").Find("td").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find("div").Eq(0).Find("a").Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + strings.Split(url, "..")[1]

		img, exists := s.Find("div").Eq(0).Find("a").Find("img").Attr("src")
		if !exists {
			return
		}
		if strings.Contains(img, "..") {
			img = strings.Split(img, "..")[1]
			img = cloud.GetURLFromCloud(info.FireStoreDir+img, info.LinkPrefix+img)
		} else {
			img = cloud.GetURLFromCloud(info.FireStoreDir+strings.Split(img, "://")[1], img)
		}

		var soldOut = false
		var nameIndex = 1

		_, exists = s.Find("div").Eq(1).Find("img").Attr("src")
		if exists {
			soldOut = true
			nameIndex = 2
		}

		name1 := util.ToUTF8(s.Find("div").Eq(nameIndex).Find("a").Text())

		name2 := ""

		price := util.ToUTF8(s.Find("b").Text())
		price = strings.TrimSpace(price)

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	s.CacheMap[query] = SearchCache{time.Now(), results}

	return results
}

// GetNewArrivals is an exported method of Crawler by HobbyGameMall
func (s HobbyGameMall) GetNewArrivals() []NewArrival {
	var results []NewArrival
	return results
}
