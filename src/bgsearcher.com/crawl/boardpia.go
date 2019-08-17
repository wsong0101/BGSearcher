package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"bgsearcher.com/util"
)

// Boardpia is a struct for shop
type Boardpia struct {
	Info ShopInfo
}

// GetSearchResults is an exported method of Crawler
func (s Boardpia) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	resp, err := http.Get(info.QueryURL + url.QueryEscape(util.ToEUCKR(query)))
	if err != nil {
		log.Printf("Boardpia: Failed to get page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Boardpia: Failed to read response")
		return results
	}

	doc.Find(".main_text").Eq(1).Find("table").Eq(1).Find("table").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find("a").Attr("href")
		if !exists {
			return
		}

		img, exists := s.Find("img").Attr("src")
		if !exists {
			return
		}

		name1 := s.Find("font.mall_product").Eq(0).Text()
		if name1 == "" {
			return
		}

		name2 := s.Find("font.mall_product").Eq(1).Text()
		name2 = strings.Split(name2, "\n")[0]

		price := s.Find("font.mall_product").Eq(1).Find("b").Text()

		var soldOut = false
		if price == util.ToEUCKR("품절") {
			soldOut = true
			price = ""
		}

		results = append(results, SearchResult{
			info.Name, url, img, util.ToUTF8(name1), util.ToUTF8(name2), util.ToUTF8(price), soldOut})
	})

	return results
}
