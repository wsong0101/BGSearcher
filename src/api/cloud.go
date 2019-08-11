package api

import (
	"context"
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

const bucketName = "bgsearcher"
const projectID = "sublime-etching-249504"
const prefix = "https://storage.cloud.google.com/" + bucketName + "/"

var fileMap map[string]bool
var hitsMap map[string]int64
var bucket *storage.BucketHandle
var collection *firestore.CollectionRef

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

	// load query hits
	hitsMap = make(map[string]int64)
	storeClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create firestore client: %v", err)
		return
	}

	collection = storeClient.Collection("history")

	iter := collection.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to get document from collection: %v", err)
			return
		}
		var query = doc.Data()["query"].(string)
		var hits = doc.Data()["hits"].(int64)
		hitsMap[query] = hits
	}
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

// IncreaseHitsCount increases hits count on firestore and cache
func IncreaseHitsCount(query string) {
	if val, exists := hitsMap[query]; exists {
		hitsMap[query] = val + 1

		go func(query string, hits int64) {
			ctx := context.Background()
			_, err := collection.Doc(query).Set(ctx, map[string]interface{}{
				"hits": hits,
			}, firestore.MergeAll)
			if err != nil {
				log.Printf("Failed to update document: %s, %d", query, hits)
			}
		}(query, val+1)

		return
	}

	// new query!
	go func(query string) {
		hitsMap[query] = 1
		ctx := context.Background()
		_, err := collection.Doc(query).Set(ctx, map[string]interface{}{
			"query": query,
			"hits":  1,
		})

		if err != nil {
			log.Printf("Failed to add on document: %s", query)
		}
	}(query)
}
