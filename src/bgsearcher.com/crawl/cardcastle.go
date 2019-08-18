package crawl

import (
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// CardCastle is a struct for shop
type CardCastle struct {
	Info ShopInfo
}

// GetSearchResults is an exported method of Crawler
func (s CardCastle) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	resp, err := http.Get(info.QueryURL + url.QueryEscape(query))
	if err != nil {
		log.Printf("CardCastle: Failed to get page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("CardCastle: Failed to read response")
		return results
	}

	doc.Find(".prdList").Find(".box").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find("a").Eq(0).Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + url

		img, exists := s.Find("a").Eq(0).Find("img").Attr("src")
		if !exists {
			return
		}
		img = "http:" + img

		var soldOut = false
		_, exists = s.Find(".status").Find(".icon").Find("img").Attr("src")
		if exists {
			soldOut = true
		}

		name1 := ""

		name2 := s.Find(".name").Find("span").Text()

		price := s.Find("strong").Text()

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})
	return results
}

// GetNewArrivals is an exported method of Crawler by CardCastle
func (s CardCastle) GetNewArrivals() []NewArrival {
	var info = &(s.Info)
	var results []NewArrival

	log.Println(info)

	return results
}
