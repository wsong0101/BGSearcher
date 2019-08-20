package crawl

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// Coupang is a struct for shop
type Coupang struct {
	Info ShopInfo
}

// GetShopInfo returns the shop's info
func (s Coupang) GetShopInfo() ShopInfo {
	return s.Info
}

var prevCookie string

// GetSearchResults is an exported method of Crawler
func (s Coupang) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	req, err := http.NewRequest("GET", "https://www.coupang.com/np/goldbox", nil)
	if err != nil {
		log.Printf("Coupang: Failed to make request")
		return results
	}

	req.AddCookie(&http.Cookie{Name: "sid", Value: "a630adde1b5140829e76fb599a1bcbb4596fbbd2"})
	req.AddCookie(&http.Cookie{Name: "PCID", Value: "227600333351734637228812"})
	req.AddCookie(&http.Cookie{Name: "overrideAbTestGroup", Value: "%5B%5D"})
	req.AddCookie(&http.Cookie{Name: "FUN", Value: "{'search':[{'reqUrl':'/search.pang','isValid':true}]}"})

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Coupang: Failed to get page")
		return results
	}
	defer resp.Body.Close()

	prevCookie = resp.Header.Get("Set-Cookie")
	log.Println(prevCookie)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Coupang: Failed to read response")
		return results
	}

	log.Println(doc.Text())
	doc.Find(".search-product").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find("a").Eq(0).Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + url

		img, exists := s.Find(".search-product-wrap-img").Attr("src")
		if !exists {
			return
		}
		img = "https:" + img

		var soldOut = false
		if s.Find(".out-of-stock").Text() != "" {
			soldOut = true
		}

		name1 := s.Find(".name").Text()
		name2 := ""

		price := s.Find(".price-value").Text()

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	return results
}

// GetNewArrivals is an exported method of Crawler by Weefun
func (s Coupang) GetNewArrivals() []NewArrival {
	var results []NewArrival
	return results
}
