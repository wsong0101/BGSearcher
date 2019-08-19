package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"bgsearcher.com/cloud"
	"github.com/PuerkitoBio/goquery"
)

// BoardgameMall is a struct for shop
type BoardgameMall struct {
	Info ShopInfo
}

// GetSearchResults is an exported method of Crawler
func (s BoardgameMall) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	resp, err := http.Get(info.QueryURL + url.QueryEscape(query))
	if err != nil {
		log.Printf("BoardgameMall: Failed to get page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("BoardgameMall: Failed to read response")
		return results
	}

	doc.Find(".goods_list_cont").Find(".item_cont").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find(".item_photo_box").Find("a").Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + strings.Split(url, "..")[1]

		img, exists := s.Find(".item_photo_box").Find("img").Attr("src")
		if !exists {
			return
		}
		img = cloud.GetURLFromCloud(info.FireStoreDir+img, info.LinkPrefix+img)

		var soldOut = false

		name1 := s.Find(".item_info_cont").Find(".item_name").Text()

		name2 := ""

		price := s.Find(".item_money_box").Find(".item_price").Text()

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	return results
}

// GetNewArrivals is an exported method of Crawler by BoardgameMall
func (s BoardgameMall) GetNewArrivals() []NewArrival {
	var results []NewArrival

	return results
}
