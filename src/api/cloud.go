package api

import (
	"context"
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

const bucketName = "bgsearcher"
const prefix = "https://storage.cloud.google.com/" + bucketName + "/"

var fileMap map[string]bool
var bucket *storage.BucketHandle

// InitializeCloud initialzies gcp client and load filemap
func InitializeCloud() {
	fileMap = make(map[string]bool)
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
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
		log.Printf("File upload completed: " + path)
	}(path, origin)

	fileMap[path] = true
	return prefix + path
}
