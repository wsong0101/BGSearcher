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

// Weefun is a struct for shop
type Weefun struct {
	Info ShopInfo
}

// GetShopInfo returns the shop's info
func (s Weefun) GetShopInfo() ShopInfo {
	return s.Info
}

// UpdatePrevNewArrivals for specific shops
func (s Weefun) UpdatePrevNewArrivals(arrivals []NewArrival) {
	return
}

// GetSearchResults is an exported method of Crawler
func (s Weefun) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	req, err := http.NewRequest("GET", info.QueryURL+url.QueryEscape(util.ToEUCKR(query)), nil)
	if err != nil {
		log.Printf("Weefun: Failed to make request: %s", err)
		return results
	}

	client := &http.Client{
		Timeout: timeoutDuration,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Weefun: Failed to get page: %s", err)
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

		name1 := util.ToUTF8(s.Find(".dsc").Text())

		name2 := util.ToUTF8(s.Find(".subname").Text())

		price := util.ToUTF8(s.Find(".price").Text())
		price = strings.TrimSpace(price)

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	return results
}

// GetNewArrivals is an exported method of Crawler by Weefun
func (s Weefun) GetNewArrivals() []NewArrival {
	var info = &(s.Info)
	var results []NewArrival

	resp, err := http.Get(info.NewArrivalURL)
	if err != nil {
		log.Printf("Weefun: Failed to get new arrival page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Weefun: Failed to read arrival response")
		return results
	}

	doc.Find(".tb-left").Each(func(i int, s *goquery.Selection) {
		title := util.ToUTF8(s.Find("a").Text())
		if title == "" {
			return
		}
		regStr := "([0-9]+)년 ([0-9]+)월 ([0-9]+)일"
		if match, _ := regexp.MatchString(regStr, title); !match {
			return
		}
		r, _ := regexp.Compile(regStr)
		matches := r.FindStringSubmatch(title)
		year := matches[1]
		month := matches[2]
		day := matches[3]

		upTime, _ := time.Parse("2006-01-02", year+"-"+month+"-"+day)
		now := time.Now()
		diff := now.Sub(upTime)
		if diff.Hours() > newArrivalsLimit {
			return
		}

		link, exists := s.Find("a").Attr("href")
		if !exists {
			return
		}
		link = info.LinkPrefix + link

		resp2, err2 := http.Get(link)
		if err2 != nil {
			log.Printf("Weefun: Failed to get new arrival detail page: %s", link)
			return
		}
		defer resp2.Body.Close()

		doc2, err2 := goquery.NewDocumentFromReader(resp2.Body)
		if err2 != nil {
			log.Printf("Weefun: Failed to read arrival detail response")
			return
		}

		var searched []SearchResult
		doc2.Find(".fixed-img-collist").Find("li").Each(func(j int, ss *goquery.Selection) {
			url, exists := ss.Find("a").Attr("href")
			if !exists {
				return
			}
			url = info.LinkPrefix + url

			img, exists := ss.Find("img").Attr("src")
			if !exists {
				return
			}
			img = info.LinkPrefix + img

			name := util.ToUTF8(ss.Find("a").Children().Eq(1).Text())
			price := util.ToUTF8(ss.Children().Eq(1).Text())

			searched = append(searched, SearchResult{
				info.Name, url, img, name, "", price, false})
		})

		results = append(results, NewArrival{upTime, searched})
	})

	return results
}
