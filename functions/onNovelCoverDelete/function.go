package onNovelCoverUpload

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	vision "cloud.google.com/go/vision/apiv1"
	"context"
	firebase "firebase.google.com/go"
	"cloud.google.com/go/storage"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	app *firebase.App
	firestoreClient *firestore.Client
	bucket *storage.BucketHandle

	visionClient *vision.ImageAnnotatorClient
)

// GCSEvent is the payload of a GCS event.
type GCSEvent struct {
	Bucket         string    `json:"bucket"`
	Name           string    `json:"name"`
	Metageneration string    `json:"metageneration"`
	ResourceState  string    `json:"resourceState"`
	TimeCreated    time.Time `json:"timeCreated"`
	Updated        time.Time `json:"updated"`
}

func init() {
	// Declare a separate err variable to avoid shadowing the client variables.
	var err error

	projectID, ok := os.LookupEnv("GPC_PROJECT")
	if !ok {
		projectID = "testing-192515"
	}

	conf := &firebase.Config{
		ProjectID: projectID,
		StorageBucket: fmt.Sprintf("%s.appspot.com", projectID),
	}

	// Initialize default app
	app, err := firebase.NewApp(context.Background(), conf)
	if err != nil {
		log.Fatalf("firebase.NewApp: %v\n", err)
	}

	// Access firestore service from the default app
	firestoreClient, err = app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("app.Firestore: %v", err)
	}

	// Access storage services from the default app
	storageClient, err := app.Storage(context.Background())
	if err != nil {
		log.Fatalf("app.Storage: %v", err)
	}

	bucket, err = storageClient.DefaultBucket()
	if err != nil {
		log.Fatalf("storageClient.DefaultBucket: %v", err)
	}
}

func OnNovelCoverDelete(ctx context.Context, e GCSEvent) error {
	_, err := metadata.FromContext(ctx) // Change _ => meta
	if err != nil {
		log.Fatalf("metadata.FromContext: %v", err)
	}

	obj := bucket.Object(fmt.Sprintf("novels/%s/cover-full.jpg", e.Name[7:27]))
	_, err = obj.Attrs(ctx) // _ => objAttrs
	if err != nil {
		if err.Error() == "storage: object doesn't exist" {
			// This file is deleted, so we can properly remove `cover` from novel
			setNovelNoCover(ctx, e.Name[7:27])
			return nil
		} else {
			log.Fatalf("obj.Attrs: %v", err, e.Name[7:27])
		}
	}

	return nil
}

func setNovelNoCover(ctx context.Context, novelId string) {
	_, err := firestoreClient.Collection("novels").Doc(novelId).Set(ctx, map[string]interface{}{
		"cover": false,
	}, firestore.MergeAll)

	if err != nil {
		log.Fatalf("setNovelNoCover error: %v\n", err.Error())
	}
}
