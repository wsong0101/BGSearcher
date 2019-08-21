package crawl

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// BoardM is a struct for shop
type BoardM struct {
	Info ShopInfo
}

// GetShopInfo returns the shop's info
func (s BoardM) GetShopInfo() ShopInfo {
	return s.Info
}

// UpdatePrevNewArrivals for specific shops
func (s BoardM) UpdatePrevNewArrivals(arrivals []NewArrival) {
	return
}

// GetSearchResults is an exported method of Crawler
func (s BoardM) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	req, err := http.NewRequest("GET", info.QueryURL+url.QueryEscape(query), nil)
	if err != nil {
		log.Printf("BoardM: Failed to make request: %s", err)
		return results
	}

	client := &http.Client{
		Timeout: timeoutDuration,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("BoardM: Failed to get page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("BoardM: Failed to read response: %s", err)
		return results
	}

	doc.Find(".space").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find(".thumbnail").Find("a").Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + strings.Split(url, "..")[1]

		img, exists := s.Find(".thumbnail").Find("img").Attr("src")
		if !exists {
			return
		}
		if !strings.Contains(img, "http") {
			img = info.LinkPrefix + img
		}

		_, soldOut := s.Find(".txt").Find("img").Attr("src")

		name1 := s.Find(".txt").Find("strong").Find("b").Text()

		name2 := s.Find(".txt").Find("strong").Text()
		if name1 != "" {
			name2 = strings.Split(name2, name1)[0]
		}

		price := s.Find(".sale").Find("strong").Text()
		if price == "" {
			price = s.Find(".cost").Find("strong").Text()
		}
		if price != "" {
			price += "원"
		}

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	return results
}

// GetNewArrivals is an exported method of Crawler by BoardM
func (s BoardM) GetNewArrivals() []NewArrival {
	var info = &(s.Info)
	var results []NewArrival

	resp, err := http.Get(info.NewArrivalURL)
	if err != nil {
		log.Printf("BoardM: Failed to get new arrival page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("BoardM: Failed to read arrival response")
		return results
	}

	doc.Find(".list_category").Find("li").Each(func(i int, s *goquery.Selection) {
		splits := strings.Split(s.Find("a").Text(), " ")
		month, _ := strconv.Atoi(strings.Split(splits[0], "월")[0])
		day := strings.Split(splits[1], "일")[0]
		now := time.Now()
		year := now.Year()
		if month > int(now.Month()) {
			year--
		}

		upTime, _ := time.Parse("2006-1-02", strconv.Itoa(year)+"-"+strconv.Itoa(month)+"-"+day)
		diff := now.Sub(upTime)
		if diff.Hours() > newArrivalsLimit {
			return
		}

		link, exists := s.Find("a").Attr("href")
		if !exists {
			return
		}

		link = strings.Split(info.NewArrivalURL, "?")[0] + link

		resp2, err2 := http.Get(link)
		if err2 != nil {
			log.Printf("BoardM: Failed to get new arrival detail page")
			return
		}
		defer resp2.Body.Close()

		doc2, err2 := goquery.NewDocumentFromReader(resp2.Body)
		if err2 != nil {
			log.Printf("BoardM: Failed to read arrival detail response")
			return
		}

		var searched []SearchResult
		doc2.Find(".space").Each(func(j int, ss *goquery.Selection) {
			url, exists := ss.Find(".thumbnail").Find("a").Attr("href")
			if !exists {
				return
			}
			url = info.LinkPrefix + strings.Split(url, "..")[1]

			img, exists := ss.Find(".thumbnail").Find("img").Attr("src")
			if !exists {
				return
			}
			if !strings.Contains(img, "http") {
				img = info.LinkPrefix + img
			}

			name := ss.Find(".txt").Find("b").Text()

			isSoldOut := false
			price := ss.Find(".sale").Text()
			if price == "" {
				price = ss.Find(".cost").Text()
				if price == "" {
					isSoldOut = true
				}
			}

			searched = append(searched, SearchResult{
				info.Name, url, img, name, "", price, isSoldOut})
		})

		results = append(results, NewArrival{upTime, searched})
	})

	return results
}
