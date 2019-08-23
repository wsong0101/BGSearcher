package crawl

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// CardCastle is a struct for shop
type CardCastle struct {
	Info     ShopInfo
	CacheMap map[string]SearchCache
}

// GetShopInfo returns the shop's info
func (s CardCastle) GetShopInfo() ShopInfo {
	return s.Info
}

// UpdatePrevNewArrivals for specific shops
func (s CardCastle) UpdatePrevNewArrivals(arrivals []NewArrival) {
	return
}

// GetSearchResults is an exported method of Crawler
func (s CardCastle) GetSearchResults(query string) []SearchResult {
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
		log.Printf("CardCastle: Failed to make request: %s", err)
		return results
	}

	client := &http.Client{
		Timeout: timeoutDuration,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("CardCastle: Failed to get page: %s", err)
		return results
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("CardCastle: Failed to read response")
		return results
	}

	doc.Find(".prdList").Find(".box").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find("a").Eq(0).Attr("href")
		if !exists {
			return
		}
		url = info.LinkPrefix + url

		img, exists := s.Find("a").Eq(0).Find("img").Attr("src")
		if !exists {
			return
		}
		img = "http:" + img

		var soldOut = false
		_, exists = s.Find(".status").Find(".icon").Find("img").Attr("src")
		if exists {
			soldOut = true
		}

		name1 := s.Find(".name").Find("span").Text()

		name2 := ""

		price := s.Find("strong").Text()
		price = strings.TrimSpace(price)

		// CardCastle doesn't show discout price on search result..
		regStr := "product_no=([0-9]*)&"
		if match, _ := regexp.MatchString(regStr, url); match {
			r, _ := regexp.Compile(regStr)
			productNo := r.FindStringSubmatch(url)[1]

			req2, err2 := http.NewRequest("GET", "http://cardcastle.co.kr/product/image_zoom.html?product_no="+productNo, nil)
			if err2 != nil {
				log.Printf("CardCastle: Failed to make detail request: %s", err2)
				return
			}
			resp2, err2 := client.Do(req2)
			if err2 != nil {
				log.Printf("CardCastle: Failed to get detail page: %s", err2)
				return
			}
			defer resp2.Body.Close()

			doc2, err2 := goquery.NewDocumentFromReader(resp2.Body)
			if err2 != nil {
				log.Printf("CardCastle: Failed to read detail response")
				return
			}
			salePrice := doc2.Find("#span_product_price_sale").Text()
			if salePrice != "" {
				price = strings.TrimSpace(strings.Split(salePrice, " ")[0])
			}
		}

		results = append(results, SearchResult{
			info.Name, url, img, name1, name2, price, soldOut})
	})

	s.CacheMap[query] = SearchCache{time.Now(), results}

	return results
}

// GetNewArrivals is an exported method of Crawler by CardCastle
func (s CardCastle) GetNewArrivals() []NewArrival {
	var results []NewArrival
	return results
}
