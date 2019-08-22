package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"bgsearcher.com/util"
)

// Boardpia is a struct for shop
type Boardpia struct {
	Info ShopInfo
}

// GetShopInfo returns the shop's info
func (s Boardpia) GetShopInfo() ShopInfo {
	return s.Info
}

// UpdatePrevNewArrivals for specific shops
func (s Boardpia) UpdatePrevNewArrivals(arrivals []NewArrival) {
	return
}

// GetSearchResults is an exported method of Crawler by Boardpia
func (s Boardpia) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	req, err := http.NewRequest("GET", info.QueryURL+url.QueryEscape(util.ToEUCKR(query)), nil)
	if err != nil {
		log.Printf("Boardpia: Failed to make request: %s", err)
		return results
	}

	client := &http.Client{
		Timeout: timeoutDuration,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Boardpia: Failed to get page: %s", err)
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

		name1 := s.Find("font.mall_product").Eq(1).Text()
		name1 = strings.Split(name1, "\n")[0]

		name2 := s.Find("font.mall_product").Eq(0).Text()
		if name2 == "" {
			return
		}

		price := s.Find("font.mall_product").Eq(1).Find("b").Text()

		var soldOut = false
		if price == util.ToEUCKR("품절") {
			soldOut = true
			price = ""
		}
		price = strings.TrimSpace(price)

		results = append(results, SearchResult{
			info.Name, url, img, util.ToUTF8(name1), util.ToUTF8(name2), util.ToUTF8(price), soldOut})
	})

	return results
}

// GetNewArrivals is an exported method of Crawler by Boardpia
func (s Boardpia) GetNewArrivals() []NewArrival {
	var info = &(s.Info)
	var results []NewArrival

	resp, err := http.Get(info.NewArrivalURL)
	if err != nil {
		log.Printf("Boardpia: Failed to get new arrival page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Boardpia: Failed to read arrival response")
		return results
	}

	doc.Find(".small_text").Each(func(i int, s *goquery.Selection) {
		upTime, _ := time.Parse("2006-01-02", s.Text())
		now := time.Now()
		diff := now.Sub(upTime)
		if diff.Hours() > newArrivalsLimit {
			return
		}

		title := util.ToUTF8(s.Siblings().Find("a").Text())
		if !strings.Contains(title, "입고 상품") {
			return
		}

		link, exists := s.Siblings().Find("a").Attr("href")
		if !exists {
			return
		}
		link = info.LinkPrefix + link

		resp2, err2 := http.Get(link)
		if err2 != nil {
			log.Printf("Boardpia: Failed to get new arrival detail page")
			return
		}
		defer resp2.Body.Close()

		doc2, err2 := goquery.NewDocumentFromReader(resp2.Body)
		if err2 != nil {
			log.Printf("Boardpia: Failed to read arrival detail response")
			return
		}

		var searched []SearchResult
		doc2.Find(".main_text").Eq(2).Children().Children().Children().Each(func(j int, ss *goquery.Selection) {
			url, exists := ss.Find("a").Attr("href")
			if !exists {
				return
			}

			img, exists := ss.Find("img").Attr("src")
			if !exists {
				return
			}

			name := ss.Find(".mall_product").Find("b").Eq(0).Text()
			name = util.ToUTF8(name)

			// name2 := ss.Find(".mall_product").Find("font").Find("font").Eq(0).Text()
			name2 := ""

			price := ss.Find(".mall_product").Find("b").Eq(1).Text()
			price = util.ToUTF8(price)
			price = strings.TrimSpace(price)

			isSoldOut := false
			if price == "품절" {
				price = ""
				isSoldOut = true
			}

			searched = append(searched, SearchResult{
				info.Name, url, img, name, name2, price, isSoldOut})
		})

		results = append(results, NewArrival{upTime, searched})
	})

	return results
}
