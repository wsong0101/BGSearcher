package cloud

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

const bucketName = "bgsearcher"
const projectID = "sublime-etching-249504"
const prefix = "https://storage.cloud.google.com/" + bucketName + "/"

var fileMap map[string]bool

var bucket *storage.BucketHandle
var collection *firestore.CollectionRef
var naCollection *firestore.CollectionRef

// InitializeCloud initialzies gcp client and load filemap
func InitializeCloud() {
	ctx := context.Background()

	// load uploaded files
	fileMap = make(map[string]bool)
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return
	}
	bucket = client.Bucket(bucketName)
	it := bucket.Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Failed to iterate object in bucket")
		}
		fileMap[attrs.Name] = true
	}

	storeClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create firestore client: %v", err)
		return
	}

	collection = storeClient.Collection("query")
	naCollection = storeClient.Collection("new-arrivals")
}

// GetURLFromCloud returns clound url, upload if no exist
func GetURLFromCloud(path string, origin string) string {
	if _, exists := fileMap[path]; exists {
		return prefix + path
	}

	go func(path string, origin string) {
		if imgRest, err := http.Get(origin); err != nil {
			log.Printf("Failed to download image from: %s", "imgpath")
		} else {
			defer imgRest.Body.Close()

			ctx := context.Background()
			wc := bucket.Object(path).NewWriter(ctx)
			if _, err = io.Copy(wc, imgRest.Body); err != nil {
				log.Printf("Failed to upload image: %v, %s", err, path)
			}

			if err := wc.Close(); err != nil {
				log.Printf("Failed to close bucket writer: %v, %s", err, path)
			}
		}
	}(path, origin)

	fileMap[path] = true
	return prefix + path
}

// SaveNewArrivals saves json string data to firestore
func SaveNewArrivals(data string) {
	ctx := context.Background()
	_, err := naCollection.Doc("latest").Set(ctx, map[string]interface{}{
		"data": data,
	})

	if err != nil {
		log.Printf("SaveNewArrivals: Failed to add on document")
	}
}

// LoadNewArrivals loades json string data from firestore
func LoadNewArrivals() string {
	ctx := context.Background()
	if result, err := naCollection.Doc("latest").Get(ctx); err == nil {
		data := result.Data()["data"]
		if s, ok := data.(string); ok {
			return s
		}
	}
	return ""
}

// AddQuery adds query and timestamp to today's document
func AddQuery(query string, now time.Time) {
	ctx := context.Background()
	if _, err := collection.Doc(now.Format("2006-01")).Collection("queries").NewDoc().Set(ctx, map[string]interface{}{
		"query":     query,
		"timestamp": now,
	}); err != nil {
		log.Printf("AddQuery: Failed to add query. query=%s", query)
	}
}

// GetQueryRange returns queries entered from 'from' to 'to'
func GetQueryRange(from time.Time, to time.Time) []string {
	ctx := context.Background()

	queryFrom := collection.Doc(from.Format("2006-01")).Collection("queries").Where("timestamp", ">=", from).Where("timestamp", "<=", to)
	iterFrom := queryFrom.Documents(ctx)

	var results []string
	for {
		doc, err := iterFrom.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("GetQueryRange: Failed. e=%s, from=%s, to=%s", err, from, to)
			break
		}
		if doc.Data()["query"] != nil {
			results = append(results, doc.Data()["query"].(string))
		}
	}

	if from.Month() == to.Month() {
		return results
	}

	queryTo := collection.Doc(from.Format("2006-01")).Collection("queries").Where("timestamp", ">=", from).Where("timestamp", "<=", to)
	iterTo := queryTo.Documents(ctx)

	for {
		doc, err := iterTo.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("GetQueryRange: Failed. e=%s, from=%s, to=%s", err, from, to)
			break
		}
		if doc.Data()["query"] != nil {
			results = append(results, doc.Data()["query"].(string))
		}
	}

	return results
}
