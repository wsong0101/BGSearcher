package crawl

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"bgsearcher.com/cloud"
)

// DevilDice is a struct for shop
type DevilDice struct {
	Info ShopInfo
}

// GetShopInfo returns the shop's info
func (s DevilDice) GetShopInfo() ShopInfo {
	return s.Info
}

// GetSearchResults is an exported method of Crawler
func (s DevilDice) GetSearchResults(query string) []SearchResult {
	var info = &(s.Info)
	var results []SearchResult

	resp, err := http.Get(info.QueryURL + url.QueryEscape(query))
	if err != nil {
		log.Printf("DevilDice: Failed to get page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("DevilDice: Failed to read response")
		return results
	}

	doc.Find(".goods-list").Find(".space").Each(func(i int, s *goquery.Selection) {
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
		img = cloud.GetURLFromCloud(info.FireStoreDir+img, info.LinkPrefix+img)

		var soldOut = false

		name1 := s.Find(".txt").Find("strong").Eq(0).Text()

		name2 := ""

		price := s.Find(".price").Find(".cost").Text()

		if price == "품절" {
			price = ""
			soldOut = true
		}

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	return results
}

// GetNewArrivals is an exported method of Crawler by DevilDice
func (s DevilDice) GetNewArrivals() []NewArrival {
	var info = &(s.Info)
	var results []NewArrival

	resp, err := http.Get(info.NewArrivalURL)
	if err != nil {
		log.Printf("DevilDice: Failed to get new arrival page")
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("DevilDice: Failed to read arrival response")
		return results
	}

	title := doc.Find(".cg-main").Find("h2").Eq(0).Text()
	regStr := "\\[.*([0-9]+.+[0-9]+).*\\]"
	if match, _ := regexp.MatchString(regStr, title); !match {
		log.Printf("DevilDice: Failed to date from title: %s", title)
		return results
	}
	r, _ := regexp.Compile(regStr)
	splits := strings.Split(r.FindStringSubmatch(title)[1], ".")
	month, _ := strconv.Atoi(splits[0])
	day := splits[1]
	now := time.Now()
	year := now.Year()
	if month > int(now.Month()) {
		year--
	}

	upTime, _ := time.Parse("2006-1-02", strconv.Itoa(year)+"-"+strconv.Itoa(month)+"-"+day)
	diff := now.Sub(upTime)
	if diff.Hours() > newArrivalsLimit {
		return results
	}

	var searched []SearchResult
	doc.Find(".space").Each(func(i int, s *goquery.Selection) {
		url, exists := s.Find(".thumbnail").Find("a").Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + strings.Split(url, "..")[1]

		img, exists := s.Find(".thumbnail").Find("img").Attr("src")
		if !exists {
			return
		}
		img = cloud.GetURLFromCloud(info.FireStoreDir+img, info.LinkPrefix+img)

		isSoldOut := false
		name := s.Find(".txt").Find("strong").Text()
		price := s.Find(".cost").Text()
		if price == "품절" {
			price = ""
			isSoldOut = true
		}

		searched = append(searched, SearchResult{
			info.Name, url, img, name, "", price, isSoldOut})
	})

	results = append(results, NewArrival{upTime, searched})
	return results
}
