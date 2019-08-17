package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"bgsearcher.com/cloud"
	"bgsearcher.com/util"
	"github.com/PuerkitoBio/goquery"
)

// HobbyGameMall is a struct for shop
type HobbyGameMall struct {
	Info ShopInfo
}

// GetSearchResults is an exported method of Crawler
func (s HobbyGameMall) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	resp, err := http.Get(info.QueryURL + url.QueryEscape(util.ToEUCKR(query)))
	if err != nil {
		log.Printf("HobbyGameMall: Failed to get page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("HobbyGameMall: Failed to read response")
		return results
	}

	doc.Find(".indiv").Find("table").Find("table").Find("td").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find("div").Eq(0).Find("a").Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + strings.Split(url, "..")[1]

		img, exists := s.Find("div").Eq(0).Find("a").Find("img").Attr("src")
		if !exists {
			return
		}
		img = strings.Split(img, "..")[1]

		img = cloud.GetURLFromCloud(info.FireStoreDir+img, info.LinkPrefix+img)

		var soldOut = false
		var nameIndex = 1

		_, exists = s.Find("div").Eq(1).Find("img").Attr("src")
		if exists {
			soldOut = true
			nameIndex = 2
		}

		name1 := ""

		name2 := util.ToUTF8(s.Find("div").Eq(nameIndex).Find("a").Text())

		price := util.ToUTF8(s.Find("b").Text())

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	return results
}
