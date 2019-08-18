package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// BoardM is a struct for shop
type BoardM struct {
	Info ShopInfo
}

// GetSearchResults is an exported method of Crawler
func (s BoardM) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	resp, err := http.Get(info.QueryURL + url.QueryEscape(query))
	if err != nil {
		log.Printf("BoardM: Failed to get page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("BoardM: Failed to read response")
		return results
	}

	doc.Find(".space").Each(func(i int, s *goquery.Selection) {
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
		if !strings.Contains(img, "http") {
			img = info.LinkPrefix + img
		}

		_, soldOut := s.Find(".txt").Find("img").Attr("src")

		name1 := s.Find(".txt").Find("strong").Text()

		name2 := s.Find(".txt").Find("strong").Find("b").Text()
		if name2 != "" {
			name1 = strings.Split(name1, name2)[0]
		}

		price := s.Find(".sale").Find("strong").Text()
		if price == "" {
			price = s.Find(".cost").Find("strong").Text()
		}
		if price != "" {
			price += "원"
		}

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	return results
}

// GetNewArrivals is an exported method of Crawler by BoardM
func (s BoardM) GetNewArrivals() []NewArrival {
	var info = &(s.Info)
	var results []NewArrival

	log.Println(info)

	return results
}
