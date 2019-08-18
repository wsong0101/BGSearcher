package crawl

import (
	"log"
	"net/http"
	"net/url"

	"bgsearcher.com/util"
	"github.com/PuerkitoBio/goquery"
)

// Weefun is a struct for shop
type Weefun struct {
	Info ShopInfo
}

// GetSearchResults is an exported method of Crawler
func (s Weefun) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	resp, err := http.Get(info.QueryURL + url.QueryEscape(util.ToEUCKR(query)))
	if err != nil {
		log.Printf("Weefun: Failed to get page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Weefun: Failed to read response")
		return results
	}

	doc.Find(".prd-list").Find(".tb-center").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find(".thumb").Find("a").Eq(0).Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + url

		img, exists := s.Find(".thumb").Find("a").Eq(0).Find("img").Attr("src")
		if !exists {
			return
		}
		img = info.LinkPrefix + img

		var soldOut = false
		if s.Find(".soldout").Text() != "" {
			soldOut = true
		}

		name1 := util.ToUTF8(s.Find(".subname").Text())

		name2 := util.ToUTF8(s.Find(".dsc").Text())

		price := util.ToUTF8(s.Find(".price").Text())

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	return results
}

// GetNewArrivals is an exported method of Crawler by Weefun
func (s Weefun) GetNewArrivals() []NewArrival {
	var info = &(s.Info)
	var results []NewArrival

	log.Println(info)

	return results
}
