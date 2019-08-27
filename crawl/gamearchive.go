package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"bgsearcher.com/cloud"
)

// GameArchive is a struct for shop
type GameArchive struct {
	Info     ShopInfo
	CacheMap map[string]SearchCache
}

// GetShopInfo returns the shop's info
func (s GameArchive) GetShopInfo() ShopInfo {
	return s.Info
}

// UpdatePrevNewArrivals for specific shops
func (s GameArchive) UpdatePrevNewArrivals(arrivals []NewArrival) {
	for i := 0; i < len(arrivals); i++ {
		result := arrivals[i].Results
		if result == nil || len(result) <= 0 {
			return
		}
		if result[0].Company == s.Info.Name {
			previousGameArcNewArrival = arrivals[i]
			return
		}
	}
}

// GetSearchResults is an exported method of Crawler
func (s GameArchive) GetSearchResults(query string) []SearchResult {
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
		log.Printf("GameArchive: Failed to make request: %s", err)
		return results
	}

	client := &http.Client{
		Timeout: timeoutDuration,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("GameArchive: Failed to get page: %s", err)
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("GameArchive: Failed to read response")
		return results
	}

	doc.Find(".list").Find("li").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find(".thumbnail").Find("a").Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + strings.Split(url, "..")[1]

		img, exists := s.Find(".thumbnail").Find("img").Attr("src")
		if !exists {
			return
		}
		img = cloud.GetURLFromCloud(info.FireStoreDir+img, info.LinkPrefix+img)

		var soldOut = false
		_, exists = s.Find(".txt").Find("img").Attr("src")
		if exists {
			soldOut = true
		}

		name1 := s.Find(".txt").Find("strong").Text()

		name2 := s.Find(".txt").Find("em").Text()

		price := s.Find(".cost").Find("strong").Text() + "원"
		price = strings.TrimSpace(price)

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	s.CacheMap[query] = SearchCache{time.Now(), results}

	return results
}

var previousGameArcNewArrival NewArrival

// GetNewArrivals is an exported method of Crawler by GameArchive
func (s GameArchive) GetNewArrivals() []NewArrival {
	var info = &(s.Info)
	var results []NewArrival
	var searched []SearchResult

	resp, err := http.Get(info.NewArrivalURL)
	if err != nil {
		log.Printf("GameArchive: Failed to get new arrival page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("GameArchive: Failed to read arrival response")
		return results
	}

	doc.Find(".space").Each(func(i int, s *goquery.Selection) {
		url, exists := s.Find(".thumbnail").Find("a").Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + strings.Split(url, "..")[1]

		img, exists := s.Find(".thumbnail").Find("img").Attr("src")
		if !exists {
			return
		}
		img = cloud.GetURLFromCloud(info.FireStoreDir+img, info.LinkPrefix+img)

		var soldOut = false
		_, exists = s.Find(".txt").Find("img").Attr("src")
		if exists {
			soldOut = true
		}

		name := s.Find(".txt").Find("strong").Text()

		price := s.Find(".cost").Find("strong").Text() + "원"
		price = strings.TrimSpace(price)

		searched = append(searched, SearchResult{
			info.Name, url, img, name, "", price, soldOut})
	})

	if isEqualSearchResults(previousGameArcNewArrival.Results, searched) {
		results = append(results, previousGameArcNewArrival)
		log.Println("GameArchive: equal result. no change.")
	} else {
		now := time.Now()
		upTime, _ := time.Parse("2006-01-02", now.Format("2006-01-02"))
		var newArrival = NewArrival{upTime, searched}
		results = append(results, newArrival)
		previousGameArcNewArrival = newArrival
	}

	return results
}
