package api

import (
	"sync"
	"time"

	"bgsearcher.com/cloud"
	"bgsearcher.com/crawl"
)

// Crawlers is the arry for shops' information
var Crawlers = []crawl.Crawler{
	crawl.BMarket{
		Info: crawl.ShopInfo{
			QueryURL:      "http://shopping.boardlife.co.kr/list.php?&action=search&search01=",
			Name:          "비마켓",
			URL:           "http://shopping.boardlife.co.kr",
			LinkPrefix:    "http://shopping.boardlife.co.kr/",
			FireStoreDir:  "bmarket",
			NewArrivalURL: "http://shopping.boardlife.co.kr/html_file.php?file=new_instock.html",
		},
	},
	crawl.BoardgameMall{
		Info: crawl.ShopInfo{
			QueryURL:     "http://www.boardgamemall.co.kr/goods/goods_search.php?keyword=",
			Name:         "보드게임몰",
			URL:          "http://boardgamemall.co.kr",
			LinkPrefix:   "http://boardgamemall.co.kr",
			FireStoreDir: "boardgamemall",
		},
	},
	crawl.BoardM{
		Info: crawl.ShopInfo{
			QueryURL:      "http://www.boardm.co.kr/goods/goods_search.php?keyword=",
			Name:          "보드엠",
			URL:           "http://www.boardm.co.kr",
			LinkPrefix:    "http://www.boardm.co.kr",
			FireStoreDir:  "boardm",
			NewArrivalURL: "http://www.boardm.co.kr/goods/goods_list.php?cateCd=024",
		},
	},
	crawl.Boardpia{
		Info: crawl.ShopInfo{
			QueryURL:      "http://boardpia.co.kr/mall/product_list.html?search=",
			Name:          "보드피아",
			URL:           "http://boardpia.co.kr",
			LinkPrefix:    "http://boardpia.co.kr/mall/",
			FireStoreDir:  "boardpia",
			NewArrivalURL: "http://boardpia.co.kr/mall/mall_notice.html",
		},
	},
	crawl.CardCastle{
		Info: crawl.ShopInfo{
			QueryURL:     "http://cardcastle.co.kr/product/search.html?category_no=1091&keyword=",
			Name:         "카드캐슬",
			URL:          "http://www.cardcastle.co.kr",
			LinkPrefix:   "http://www.cardcastle.co.kr",
			FireStoreDir: "cardcastle",
		},
	},
	crawl.DevilDice{
		Info: crawl.ShopInfo{
			QueryURL:      "http://devildice.co.kr/goods/goods_search.php?keyword=",
			Name:          "데블다이스",
			URL:           "http://devildice.co.kr",
			LinkPrefix:    "http://devildice.co.kr",
			FireStoreDir:  "devildice",
			NewArrivalURL: "http://devildice.co.kr/goods/goods_main.php?sno=3",
		},
	},
	crawl.DiveDice{
		Info: crawl.ShopInfo{
			QueryURL:     "https://www.divedice.com/_proc/prd/prd_list.php",
			Name:         "다이브다이스",
			URL:          "https://www.divedice.com",
			LinkPrefix:   "https://www.divedice.com/site/game/",
			FireStoreDir: "divedice",
		},
	},
	crawl.GameArchive{
		Info: crawl.ShopInfo{
			QueryURL:      "http://gamearc.co.kr/goods/goods_search.php?keyword=",
			Name:          "게임아카이브",
			URL:           "http://www.gamearc.co.kr",
			LinkPrefix:    "http://www.gamearc.co.kr",
			FireStoreDir:  "gamearchive",
			NewArrivalURL: "http://gamearc.co.kr/goods/goods_main.php?sno=8",
		},
	},
	crawl.HobbyGameMall{
		Info: crawl.ShopInfo{
			QueryURL:     "http://www.hobbygamemall.com/shop/goods/goods_search.php?searched=Y&skey=all&sword=&sword=",
			Name:         "하비게임몰",
			URL:          "http://www.hobbygamemall.com",
			LinkPrefix:   "http://www.hobbygamemall.com/shop",
			FireStoreDir: "hobbygame",
		},
	},
	crawl.PopconeEdu{
		Info: crawl.ShopInfo{
			QueryURL:     "http://www.popcone.co.kr/shop/goods/goods_search.php?disp_type=gallery&searched=Y&skey=all&cate[0]=002&sword=",
			Name:         "팝콘에듀",
			URL:          "http://www.popcone.co.kr",
			LinkPrefix:   "http://www.popcone.co.kr/shop",
			FireStoreDir: "popconeedu",
		},
	},
	crawl.Weefun{
		Info: crawl.ShopInfo{
			QueryURL:      "http://weefun.co.kr/shop/shopbrand.html?search&page=1&sort=brandname&prize1=",
			Name:          "위펀",
			URL:           "http://www.weefun.co.kr",
			LinkPrefix:    "http://www.weefun.co.kr",
			FireStoreDir:  "weefun",
			NewArrivalURL: "http://www.weefun.co.kr/board/board.html?code=weefun_board9",
		},
	},
	crawl.Coupang{
		Info: crawl.ShopInfo{
			QueryURL:     "https://www.coupang.com/np/search?component=332130&eventCategory=SRP&sorter=scoreDesc&filterType=rocket,&listSize=72&isPriceRange=false&rating=0&page=1&rocketAll=false&q=",
			Name:         "쿠팡-로켓배송",
			URL:          "https://www.coupang.com",
			LinkPrefix:   "https://www.coupang.com",
			FireStoreDir: "coupang",
		},
	},
}

// Search returns a slice of SearchResult
func Search(query string) []crawl.SearchResult {
	size := len(Crawlers)
	wg := sync.WaitGroup{}
	wg.Add(size)
	ch := make(chan []crawl.SearchResult, size)

	for i := 0; i < size; i++ {
		var crawler = Crawlers[i]

		go func(ch chan []crawl.SearchResult, wg *sync.WaitGroup, crawler crawl.Crawler, query string) {
			defer wg.Done()
			results := crawler.GetSearchResults(query)
			ch <- results
		}(ch, &wg, crawler, query)
	}

	wg.Wait()
	close(ch)

	var results []crawl.SearchResult
	for {
		if result, success := <-ch; success {
			results = append(results, result...)
		} else {
			cloud.IncreaseHitsCount(query)
			return results
		}
	}
}

var newArrivals []crawl.NewArrival

// UpdateNewArrivals runs repeatedly to crawl sites' new arrivals
func UpdateNewArrivals(period time.Duration) {
	for {
		size := len(Crawlers)
		wg := sync.WaitGroup{}
		wg.Add(size)
		ch := make(chan []crawl.NewArrival, size)

		for i := 0; i < size; i++ {
			var crawler = Crawlers[i]

			go func(ch chan []crawl.NewArrival, wg *sync.WaitGroup, crawler crawl.Crawler) {
				defer wg.Done()
				result := crawler.GetNewArrivals()
				if len(result) <= 0 {
					return
				}
				ch <- result
			}(ch, &wg, crawler)
		}

		wg.Wait()
		close(ch)

		newArrivals = newArrivals[:0]
		for {
			if result, success := <-ch; success {
				newArrivals = append(newArrivals, result...)
			} else {
				break
			}
		}

		<-time.After(period)
	}
}

// GetNewArrivalsFromCache returns cached new arrivals data
func GetNewArrivalsFromCache() []crawl.NewArrival {
	return newArrivals
}

// GetShopInfos returns all shop's infos
func GetShopInfos() []crawl.ShopInfo {
	var shopInfos []crawl.ShopInfo
	for i := 0; i < len(Crawlers); i++ {
		crawler := Crawlers[i]
		shopInfos = append(shopInfos, crawler.GetShopInfo())
	}
	return shopInfos
}
