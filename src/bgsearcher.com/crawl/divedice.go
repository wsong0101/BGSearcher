package crawl

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DiveDice is a struct for shop
type DiveDice struct {
	Info ShopInfo
}

// ddResponse is a struct form DiveDice's json search response
type ddResponse struct {
	Total    string
	TotPage  int
	PageList string
	Offset   int
	SQL      string
	Pagenum  int
	HTML     string
}

// GetSearchResults is an exported method of Crawler
func (s DiveDice) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	resp, err := http.PostForm(info.QueryURL, url.Values{
		"top_name": {query},
	})
	if err != nil {
		log.Printf("DiveDice: Failed to get page")
		return results
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("DiveDice: Failed to read response")
		return results
	}

	var ddResp ddResponse
	err = json.Unmarshal([]byte(respBody), &ddResp)
	if err != nil {
		log.Printf("DiveDice: Failed to unmarshal response")
		return results
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(ddResp.HTML))
	if err != nil {
		log.Printf("DiveDice: Failed to read from response")
		return results
	}

	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find(".thum").Children().Eq(1).Attr("href")
		if !exists {
		}
		url = info.LinkPrefix + url

		img, exists := s.Find(".thum").Find("img").Attr("src")
		if !exists {
			return
		}

		name1 := s.Find("h3").Find("a").Text()

		name2 := ""

		price := s.Find(".price").Text()

		var soldOut = false
		discount := s.Find(".dc").Text()
		if discount == "품절" {
			soldOut = true
		}

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	return results
}

// GetNewArrivals is an exported method of Crawler by DiveDice
func (s DiveDice) GetNewArrivals() []NewArrival {
	var results []NewArrival
	return results
}
