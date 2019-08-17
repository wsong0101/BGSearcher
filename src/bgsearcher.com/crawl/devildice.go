package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"bgsearcher.com/cloud"
)

// DevilDice is a struct for shop
type DevilDice struct {
	Info ShopInfo
}

// GetSearchResults is an exported method of Crawler
func (s DevilDice) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	resp, err := http.Get(info.QueryURL + url.QueryEscape(query))
	if err != nil {
		log.Printf("DevilDice: Failed to get page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("DevilDice: Failed to read response")
		return results
	}

	doc.Find(".goods-list").Find(".space").Each(func(i int, s *goquery.Selection) {
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

		name1 := s.Find(".txt").Find("strong").Eq(0).Text()

		name2 := ""

		price := s.Find(".price").Find(".cost").Text()

		if price == "품절" {
			price = ""
			soldOut = true
		}

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	return results
}
