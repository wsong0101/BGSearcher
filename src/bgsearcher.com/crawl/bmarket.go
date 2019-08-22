package crawl

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"bgsearcher.com/util"
	"github.com/PuerkitoBio/goquery"
)

// BMarket is a struct for shop
type BMarket struct {
	Info     ShopInfo
	CacheMap map[string]SearchCache
}

// GetShopInfo returns the shop's info
func (s BMarket) GetShopInfo() ShopInfo {
	return s.Info
}

// UpdatePrevNewArrivals for specific shops
func (s BMarket) UpdatePrevNewArrivals(arrivals []NewArrival) {
	return
}

// GetSearchResults is an exported method of Crawler
func (s BMarket) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	if val, exists := s.CacheMap[query]; exists {
		now := time.Now()
		if now.Sub(val.SearchedTime) <= searchCacheDuration {
			return val.Results
		}
	}

	req, err := http.NewRequest("GET", info.QueryURL+url.QueryEscape(query), nil)
	if err != nil {
		log.Printf("BMarket: Failed to make request: %s", err)
		return results
	}

	client := &http.Client{
		Timeout: timeoutDuration,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("BMarket: Failed to get page: %s", err)
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
		img = info.LinkPrefix + strings.Split(img, "./")[1]

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
			price = strings.TrimSpace(price)
		}

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	s.CacheMap[query] = SearchCache{time.Now(), results}

	return results
}

// GetNewArrivals is an exported method of Crawler by BMarket
func (s BMarket) GetNewArrivals() []NewArrival {
	var info = &(s.Info)
	var results []NewArrival

	resp, err := http.Get(info.NewArrivalURL)
	if err != nil {
		log.Printf("BMarket: Failed to get new arrival page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("BMarket: Failed to read arrival response")
		return results
	}

	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		upTime, err := time.Parse("2006-01-02", s.Children().Eq(5).Text())
		if err != nil {
			return
		}
		now := time.Now()
		diff := now.Sub(upTime)
		if diff.Hours() > newArrivalsLimit {
			return
		}

		link, exists := s.Children().Eq(3).Children().Attr("href")
		if !exists {
			return
		}
		link = "http://boardlife.co.kr/" + strings.Split(link, "./")[1]

		resp2, err2 := http.Get(link)
		if err2 != nil {
			log.Printf("BMarket: Failed to get new arrival detail page")
			return
		}
		defer resp2.Body.Close()

		doc2, err2 := goquery.NewDocumentFromReader(resp2.Body)
		if err2 != nil {
			log.Printf("BMarket: Failed to read arrival detail response")
			return
		}

		var searched []SearchResult
		doc2.Find("#board_contents").Find("tbody").Each(func(i int, ss *goquery.Selection) {
			aTag := ss.Children().Eq(0).Children().Children().Children().Children().Children().Children()
			url, exists := aTag.Attr("href")
			if !exists {
				return
			}

			img, exists := aTag.Children().Attr("src")
			if !exists {
				return
			}
			img = "http://boardlife.co.kr/" + strings.Split(img, "./")[1]

			name := util.ToUTF8(ss.Find("strong").Eq(0).Text())

			price := util.ToUTF8(ss.Find("strong").Eq(1).Text())
			price = strings.TrimSpace(price)
			if !strings.Contains(price, "원") {
				price += "원"
			}

			isSoldOut := false
			if strings.Contains(price, "품절") {
				isSoldOut = true
				price = ""
			}

			searched = append(searched, SearchResult{
				info.Name, url, img, name, "", price, isSoldOut})
		})

		results = append(results, NewArrival{upTime, searched})
	})

	return results
}
