package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"bgsearcher.com/cloud"
)

// GameArchive is a struct for shop
type GameArchive struct {
	Info ShopInfo
}

// GetSearchResults is an exported method of Crawler
func (s GameArchive) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	resp, err := http.Get(info.QueryURL + url.QueryEscape(query))
	if err != nil {
		log.Printf("GameArchive: Failed to get page")
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

		name1 := s.Find(".txt").Find("em").Text()

		name2 := s.Find(".txt").Find("strong").Text()

		price := s.Find(".cost").Find("strong").Text() + "원"

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	return results
}
