package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

// SearchResult represents individual game info
type SearchResult struct {
	Company string
	Link    string
	Img     string
	Name    string
	Name2   string
	Price   string
	SoldOut bool
}

// Search returns a slice of SearchResult
func Search(query string) []SearchResult {
	var results []SearchResult
	results = append(results, searchBoardpia(query)...)
	results = append(results, searchBoardm(query)...)
	results = append(results, searchDivedice(query)...)

	return results
}

func toUTF8(s string) string {
	var bufs bytes.Buffer
	wr := transform.NewWriter(&bufs, korean.EUCKR.NewDecoder())
	wr.Write([]byte(s))
	wr.Close()

	return bufs.String()
}

func toEUCKR(s string) string {
	var bufs bytes.Buffer
	wr := transform.NewWriter(&bufs, korean.EUCKR.NewEncoder())
	wr.Write([]byte(s))
	wr.Close()

	return bufs.String()
}

func searchBoardpia(query string) []SearchResult {
	resp, err := http.Get("http://boardpia.co.kr/mall/product_list.html?search=" + url.QueryEscape(toEUCKR(query)))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	var results []SearchResult

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

		name1 := s.Find("font.mall_product").Eq(0).Text()
		if name1 == "" {
			return
		}

		name2 := s.Find("font.mall_product").Eq(1).Text()
		name2 = strings.Split(name2, "\n")[0]

		price := s.Find("font.mall_product").Eq(1).Find("b").Text()

		var soldOut = false
		if price == toEUCKR("품절") {
			soldOut = true
			price = ""
		}

		results = append(results, SearchResult{
			"보드피아", url, img, toUTF8(name1), toUTF8(name2), toUTF8(price), soldOut})
	})

	return results
}

func searchBoardm(query string) []SearchResult {
	resp, err := http.Get("http://www.boardm.co.kr/goods/goods_search.php?keyword=" + url.QueryEscape(query))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	var results []SearchResult

	doc.Find(".space").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find(".thumbnail").Find("a").Attr("href")
		if !exists {
			return
		}
		url = "http://www.boardm.co.kr" + strings.Split(url, "..")[1]

		img, exists := s.Find(".thumbnail").Find("img").Attr("src")
		if !exists {
			return
		}
		if !strings.Contains(img, "http") {
			img = "http://www.boardm.co.kr" + img
		}

		_, soldOut := s.Find(".txt").Find("img").Attr("src")

		name1 := s.Find(".txt").Find("strong").Text()

		name2 := s.Find(".txt").Find("strong").Find("b").Text()
		if name2 != "" {
			name1 = strings.Split(name1, name2)[0]
		}

		price := s.Find(".sale").Find("strong").Text()
		if price == "" {
			price = s.Find(".cost").Find("strong").Text()
		}
		if price != "" {
			price += "원"
		}

		results = append(results, SearchResult{
			"보드엠", url, img, name1, name2, price, soldOut})
	})

	return results
}

type DDResponse struct {
	Total    string
	TotPage  int
	PageList string
	Offset   int
	SQL      string
	Pagenum  int
	HTML     string
}

func searchDivedice(query string) []SearchResult {
	resp, err := http.PostForm("https://www.divedice.com/_proc/prd/prd_list.php", url.Values{
		"top_name": {query},
	})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var ddResp DDResponse
	err = json.Unmarshal([]byte(respBody), &ddResp)
	if err != nil {
		panic(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(ddResp.HTML))
	if err != nil {
		panic(err)
	}

	var results []SearchResult

	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find(".thum").Find("a").Eq(1).Attr("href")
		if !exists {
			return
		}
		url = "https://www.divedice.com/site/game/" + url

		img, exists := s.Find(".thum").Find("img").Attr("src")
		if !exists {
			return
		}

		name1 := ""

		name2 := s.Find("h3").Find("a").Text()

		price := s.Find(".price").Text()

		var soldOut = false
		discount := s.Find(".dc").Text()
		if discount == "품절" {
			soldOut = true
		}

		results = append(results, SearchResult{
			"다이브다이스", url, img, name1, name2, price, soldOut})
	})

	return results
}
