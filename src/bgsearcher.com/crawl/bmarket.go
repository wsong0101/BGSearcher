package crawl

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"bgsearcher.com/util"
	"github.com/PuerkitoBio/goquery"
)

// BMarket is a struct for shop
type BMarket struct {
	Info ShopInfo
}

// GetSearchResults is an exported method of Crawler
func (s BMarket) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	resp, err := http.Get(info.QueryURL + url.QueryEscape(query))
	if err != nil {
		log.Printf("BMarket: Failed to get page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("BMarket: Failed to read response")
		return results
	}

	doc.Find("#stabTLayer1").Find("table").Eq(0).Find("div").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find("a").Eq(0).Attr("href")
		if !exists || !strings.Contains(url, "view.php") {
			return
		}
		url = info.LinkPrefix + strings.Split(url, "./")[1]

		img, exists := s.Find("img").Eq(0).Attr("src")
		if !exists || !strings.Contains(img, "thumb") {
			img, exists = s.Find("img").Eq(1).Attr("src")
			if !exists || !strings.Contains(img, "thumb") {
				return
			}
		}
		img = strings.Split(img, "./")[1]
		img = info.LinkPrefix + img

		var soldOut = false

		var target = s.Children().Eq(0).Children().Eq(0).ChildrenFiltered("tr").Eq(1)
		name1 := util.ToUTF8(target.Find("span").Eq(0).Text())

		name2 := util.ToUTF8(target.Find("font").Eq(0).Text())

		var price = ""

		priceURL, exists := target.Find("img").Attr("src")
		if !exists {
			return
		}
		priceURL = util.ToUTF8(priceURL)
		if strings.Contains(priceURL, "품절") {
			soldOut = true
			price = ""
		} else {
			if match, _ := regexp.MatchString("=([0-9]*,*[0-9]* 원)", priceURL); !match {
				return
			}
			r, _ := regexp.Compile("=([0-9]*,*[0-9]* 원)")
			price = r.FindStringSubmatch(priceURL)[1]
		}

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	return results
}