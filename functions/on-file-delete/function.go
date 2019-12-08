package onfiledelete

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/storage"
	vision "cloud.google.com/go/vision/apiv1"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var (
	app             *firebase.App
	firestoreClient *firestore.Client
	bucket          *storage.BucketHandle

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
	ctx := context.Background()
	// Declare a separate err variable to avoid shadowing the client variables.
	var err error

	projectID, ok := os.LookupEnv("GPC_PROJECT")
	if !ok {
		projectID = "testing-192515"
	}

	// Initialize the app with a custom auth variable, limiting the server's access
	ao := map[string]interface{}{
		"uid": "my-service-worker",
	}
	conf := &firebase.Config{
		ProjectID:     projectID,
		StorageBucket: fmt.Sprintf("%s.appspot.com", projectID),
		AuthOverride:  &ao,
	}

	// Fetch the service account key JSON file contents
	//opt := option.WithCredentialsFile("firebase-adminsdk-testing-192515.json")

	// Initialize default app
	app, err := firebase.NewApp(ctx, conf/*, opt*/)
	if err != nil {
		log.Fatalf("firebase.NewApp: %v\n", err)
	}

	// Access firestore service from the default app
	firestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("app.Firestore: %v", err)
	}

	// Access storage services from the default app
	storageClient, err := app.Storage(ctx)
	if err != nil {
		log.Fatalf("app.Storage: %v", err)
	}

	bucket, err = storageClient.DefaultBucket()
	if err != nil {
		log.Fatalf("storageClient.DefaultBucket: %v", err)
	}
}

// OnNovelCoverDelete is a function executing upon deleting cover of a novel.
// usage: when a cover is deleted, updates novel to match the change
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
		}

		log.Fatalf("obj.Attrs: %v", err)
	}

	return nil
}

func setNovelNoCover(ctx context.Context, novelID string) {
	_, err := firestoreClient.Collection("novels").Doc(novelID).Set(ctx, map[string]interface{}{
		"cover": false,
	}, firestore.MergeAll)

	if err != nil {
		log.Fatalf("setNovelNoCover error: %v\n", err.Error())
	}
}
