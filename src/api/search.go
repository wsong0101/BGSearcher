package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

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
	wg := sync.WaitGroup{}
	wg.Add(7)
	ch := make(chan []SearchResult, 7)

	go searchCardcastle(ch, &wg, query)
	go searchGameArchive(ch, &wg, query)
	go searchBoardpia(ch, &wg, query)
	go searchBoardm(ch, &wg, query)
	go searchDivedice(ch, &wg, query)
	go searchPopcone(ch, &wg, query)
	go searchHobbygame(ch, &wg, query)

	wg.Wait()
	close(ch)

	var results []SearchResult
	for {
		if result, success := <-ch; success {
			results = append(results, result...)
		} else {
			IncreaseHitsCount(query)
			return results
		}
	}
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

func searchBoardpia(ch chan []SearchResult, wg *sync.WaitGroup, query string) {
	defer wg.Done()
	resp, err := http.Get("http://boardpia.co.kr/mall/product_list.html?search=" + url.QueryEscape(toEUCKR(query)))
	if err != nil {
		log.Printf("searchBoardpia: Failed to get page")
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("searchBoardpia: Failed to read response")
		return
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

	ch <- results
}

func searchBoardm(ch chan []SearchResult, wg *sync.WaitGroup, query string) {
	defer wg.Done()
	resp, err := http.Get("http://www.boardm.co.kr/goods/goods_search.php?keyword=" + url.QueryEscape(query))
	if err != nil {
		log.Printf("searchBoardm: Failed to get page")
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("searchBoardm: Failed to read response")
		return
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

	ch <- results
}

// DDResponse is a struct form DiveDice's json search response
type ddResponse struct {
	Total    string
	TotPage  int
	PageList string
	Offset   int
	SQL      string
	Pagenum  int
	HTML     string
}

func searchDivedice(ch chan []SearchResult, wg *sync.WaitGroup, query string) {
	defer wg.Done()
	resp, err := http.PostForm("https://www.divedice.com/_proc/prd/prd_list.php", url.Values{
		"top_name": {query},
	})
	if err != nil {
		log.Printf("searchDivedice: Failed to get page")
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("searchDivedice: Failed to read response")
		return
	}

	var ddResp ddResponse
	err = json.Unmarshal([]byte(respBody), &ddResp)
	if err != nil {
		log.Printf("searchDivedice: Failed to unmarshal response")
		return
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(ddResp.HTML))
	if err != nil {
		log.Printf("searchDivedice: Failed to read from response")
		return
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

	ch <- results
}

func searchPopcone(ch chan []SearchResult, wg *sync.WaitGroup, query string) {
	defer wg.Done()
	resp, err := http.Get("http://www.popcone.co.kr/shop/goods/goods_search.php?disp_type=gallery&searched=Y&skey=all&sword=" + url.QueryEscape(toEUCKR(query)))
	if err != nil {
		log.Printf("searchPopcone: Failed to get page")
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("searchPopcone: Failed to read response")
		return
	}

	var results []SearchResult

	doc.Find(".item_list_wrap").Find("td").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find(".item_thum").Find("a").Eq(1).Attr("href")
		if !exists {
			return
		}
		url = "http://www.popcone.co.kr/shop" + strings.Split(url, "..")[1]

		img, exists := s.Find(".item_thum").Find("a").Eq(1).Find("img").Attr("src")
		if !exists {
			return
		}
		if !strings.Contains(img, "http") {
			img = "http://www.popcone.co.kr/shop" + strings.Split(img, "..")[1]
		}

		name1 := ""

		name2 := toUTF8(s.Find(".i_name").Text())

		price := s.Find(".price").Text() + "원"

		var soldOut = false
		var states = ""
		s.Find(".i_state").Find("img").Each(func(i int, ss *goquery.Selection) {
			src, exists := ss.Attr("src")
			if exists {
				states += src
			}
		})
		if strings.Contains(states, "soldout") {
			soldOut = true
		}

		results = append(results, SearchResult{
			"팝콘에듀", url, img, name1, name2, price, soldOut})
	})

	ch <- results
}

func searchHobbygame(ch chan []SearchResult, wg *sync.WaitGroup, query string) {
	defer wg.Done()
	resp, err := http.Get("http://www.hobbygamemall.com/shop/goods/goods_search.php?searched=Y&skey=all&sword=&sword=" + url.QueryEscape(toEUCKR(query)))
	if err != nil {
		log.Printf("searchHobbygame: Failed to get page")
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("searchHobbygame: Failed to read response")
		return
	}

	var results []SearchResult

	doc.Find(".indiv").Find("table").Find("table").Find("td").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find("div").Eq(0).Find("a").Attr("href")
		if !exists {
			return
		}
		url = "http://www.hobbygamemall.com/shop" + strings.Split(url, "..")[1]

		img, exists := s.Find("div").Eq(0).Find("a").Find("img").Attr("src")
		if !exists {
			return
		}
		img = strings.Split(img, "..")[1]

		img = GetURLFromCloud("hobbygame"+img, "http://www.hobbygamemall.com/shop"+img)

		var soldOut = false
		var nameIndex = 1

		_, exists = s.Find("div").Eq(1).Find("img").Attr("src")
		if exists {
			soldOut = true
			nameIndex = 2
		}

		name1 := ""

		name2 := toUTF8(s.Find("div").Eq(nameIndex).Find("a").Text())

		price := toUTF8(s.Find("b").Text())

		results = append(results, SearchResult{
			"하비게임몰", url, img, name1, name2, price, soldOut})
	})

	ch <- results
}

func searchGameArchive(ch chan []SearchResult, wg *sync.WaitGroup, query string) {
	defer wg.Done()
	resp, err := http.Get("http://gamearc.co.kr/goods/goods_search.php?keyword=" + url.QueryEscape(query))
	if err != nil {
		log.Printf("searchHobbygame: Failed to get page")
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("searchHobbygame: Failed to read response")
		return
	}

	var results []SearchResult

	doc.Find(".list").Find("li").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find(".thumbnail").Find("a").Attr("href")
		if !exists {
			return
		}
		url = "http://www.gamearc.co.kr" + strings.Split(url, "..")[1]

		img, exists := s.Find(".thumbnail").Find("img").Attr("src")
		if !exists {
			return
		}
		img = GetURLFromCloud("gamearc"+img, "http://www.gamearc.co.kr"+img)

		var soldOut = false
		_, exists = s.Find(".txt").Find("img").Attr("src")
		if exists {
			soldOut = true
		}

		name1 := s.Find(".txt").Find("em").Text()

		name2 := s.Find(".txt").Find("strong").Text()

		price := s.Find(".cost").Find("strong").Text() + "원"

		results = append(results, SearchResult{
			"게임아카이브", url, img, name1, name2, price, soldOut})
	})

	ch <- results
}

func searchCardcastle(ch chan []SearchResult, wg *sync.WaitGroup, query string) {
	defer wg.Done()
	resp, err := http.Get("http://cardcastle.co.kr/product/search.html?&keyword=" + url.QueryEscape(query))
	if err != nil {
		log.Printf("searchCardcastle: Failed to get page")
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("searchCardcastle: Failed to read response")
		return
	}

	var results []SearchResult

	doc.Find(".prdList").Find(".box").Each(func(i int, s *goquery.Selection) {
		// each item
		url, exists := s.Find("a").Eq(0).Attr("href")
		if !exists {
			return
		}
		url = "http://www.cardcastle.co.kr" + url

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

		name1 := ""

		name2 := s.Find(".name").Find("span").Text()

		price := s.Find("strong").Text()

		results = append(results, SearchResult{
			"카드캐슬", url, img, name1, name2, price, soldOut})
	})

	ch <- results
}
