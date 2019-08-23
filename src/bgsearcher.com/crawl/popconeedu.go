package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bgsearcher.com/util"
	"github.com/PuerkitoBio/goquery"
)

// PopconeEdu is a struct for shop
type PopconeEdu struct {
	Info     ShopInfo
	CacheMap map[string]SearchCache
}

// GetShopInfo returns the shop's info
func (s PopconeEdu) GetShopInfo() ShopInfo {
	return s.Info
}

// UpdatePrevNewArrivals for specific shops
func (s PopconeEdu) UpdatePrevNewArrivals(arrivals []NewArrival) {
	return
}

// GetSearchResults is an exported method of Crawler
func (s PopconeEdu) GetSearchResults(query string) []SearchResult {
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
		log.Printf("PopconeEdu: Failed to make request: %s", err)
		return results
	}

	client := &http.Client{
		Timeout: timeoutDuration,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("PopconeEdu: Failed to get page: %s", err)
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("PopconeEdu: Failed to read response")
		return results
	}

	doc.Find(".item_list_wrap").Find("td").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find(".item_thum").Find("a").Eq(1).Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + strings.Split(url, "..")[1]

		img, exists := s.Find(".item_thum").Find("a").Eq(1).Find("img").Attr("src")
		if !exists {
			return
		}
		if !strings.Contains(img, "http") {
			img = info.LinkPrefix + strings.Split(img, "..")[1]
		}

		name1 := util.ToUTF8(s.Find(".i_name").Text())

		name2 := ""

		price := s.Find(".c_price").Children().Text()
		if price == "" {
			price = s.Find(".price").Text()
		}
		if price != "" {
			price = strings.TrimSpace(price) + "원"
		}

		var soldOut = false
		var states = ""
		s.Find(".i_state").Find("img").Each(func(i int, ss *goquery.Selection) {
			src, exists := ss.Attr("src")
			if exists {
				states += src
			}
		})
		if strings.Contains(states, "soldout") {
			soldOut = true
		}

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	s.CacheMap[query] = SearchCache{time.Now(), results}

	return results
}

// GetNewArrivals is an exported method of Crawler by PopconeEdu
func (s PopconeEdu) GetNewArrivals() []NewArrival {
	var results []NewArrival
	return results
}
