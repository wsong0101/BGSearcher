package ranking

import (
	"log"
	"time"

	"bgsearcher.com/cloud"
)

var topRankingCount = 20

// QueryCount is a struct for query to searched count
type QueryCount struct {
	Name  string
	Count int64
}

type rankUpdator struct {
	updateTime     time.Time
	updateDuration time.Duration
	rankPeriod     time.Duration
	countMap       map[string]int64
	topRanks       []QueryCount
}

func initUpdator(u *rankUpdator, now time.Time) {
	from := now.Add(-u.rankPeriod)
	results := cloud.GetQueryRange(from, now)

	for _, result := range results {
		add(u, result)
	}

	u.updateTime = now.Add(u.updateDuration)
}

func update(u *rankUpdator, now time.Time) {
	if !now.After(u.updateTime) {
		return
	}

	from := u.updateTime.Add(-u.rankPeriod).Add(-u.updateDuration)
	to := from.Add(u.updateDuration)
	removes := cloud.GetQueryRange(from, to)

	for _, remove := range removes {
		minus(u, remove)
	}

	u.updateTime = u.updateTime.Add(u.updateDuration)
}

func add(u *rankUpdator, query string) {
	var count int64
	if val, exists := u.countMap[query]; exists {
		u.countMap[query] = val + 1
		count = val + 1
	} else {
		u.countMap[query] = 1
		count = 1
	}

	// check if the query is already in topRanks
	updated := -1
	for i := len(u.topRanks) - 1; i >= 0; i-- {
		q := &u.topRanks[i]

		if updated != -1 {
			target := &u.topRanks[updated]
			if q.Count < target.Count {
				// rank up!
				*q, *target = *target, *q
				updated = i
				continue
			}
			break
		}

		if q.Name == query {
			q.Count = count
			updated = i
		}
	}
	if updated != -1 {
		return
	}

	// check if the query is a new topRank
	topCount := len(u.topRanks)
	if topCount < topRankingCount {
		u.topRanks = append(u.topRanks, QueryCount{
			Name:  query,
			Count: count,
		})
		return
	}

	q := &u.topRanks[len(u.topRanks)-1]
	if q.Count < count {
		*q = QueryCount{
			Name:  query,
			Count: count,
		}
	}
}

func minus(u *rankUpdator, query string) {
	var count int64
	if val, exists := u.countMap[query]; exists {
		if val < 1 {
			log.Printf("minus: zero count. q=%s", query)
			return
		}
		u.countMap[query] = val - 1
		count = val - 1
	} else {
		log.Printf("minus: no query. q=%s", query)
		return
	}

	for i := 0; i < len(u.topRanks); i++ {
		q := &u.topRanks[i]
		if q.Name == query {
			q.Count = count
			break
		}
	}
}

func remove(u *rankUpdator, query string) {
	delete(u.countMap, query)
	for i := 0; i < len(u.topRanks); i++ {
		q := &u.topRanks[i]
		if q.Name == query {
			u.topRanks = append(u.topRanks[:i], u.topRanks[i+1:]...)
			return
		}
	}
}

var monthly rankUpdator
var weekly rankUpdator
var hourly rankUpdator

// InitRanking inits ranking from cloud
func InitRanking() {
	now := time.Now()

	monthly.updateDuration = time.Duration(24 * time.Hour)
	monthly.rankPeriod = time.Duration(30 * 24 * time.Hour)
	monthly.countMap = make(map[string]int64)
	initUpdator(&monthly, now)

	weekly.updateDuration = time.Duration(24 * time.Hour)
	weekly.rankPeriod = time.Duration(7 * 24 * time.Hour)
	weekly.countMap = make(map[string]int64)
	initUpdator(&weekly, now)

	hourly.updateDuration = time.Duration(30 * time.Minute)
	hourly.rankPeriod = time.Duration(1 * time.Hour)
	hourly.countMap = make(map[string]int64)
	initUpdator(&hourly, now)
}

// AddQuery addes a count to the query
func AddQuery(query string) {
	now := time.Now()

	// upload to firestore
	cloud.AddQuery(query, now)

	add(&monthly, query)
	update(&monthly, now)

	add(&weekly, query)
	update(&weekly, now)

	add(&hourly, query)
	update(&hourly, now)
}

// RemoveQuery removes cached rank and query from cloud
func RemoveQuery(query string) {
	remove(&monthly, query)
	remove(&weekly, query)
	remove(&hourly, query)

	from := monthly.updateTime.Add(-monthly.rankPeriod).Add(-monthly.updateDuration)
	cloud.RemoveQueryRange(from, time.Now(), query)
}

// GetMonthlyRank returns monthly rank
func GetMonthlyRank() []QueryCount {
	return monthly.topRanks
}

// GetWeeklyRank returns weekly rank
func GetWeeklyRank() []QueryCount {
	return weekly.topRanks
}

// GetHourlyRank returns hourly rank
func GetHourlyRank() []QueryCount {
	return hourly.topRanks
}
