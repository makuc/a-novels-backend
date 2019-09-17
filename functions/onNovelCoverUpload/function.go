package onNovelCoverUpload

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/storage"
	vision "cloud.google.com/go/vision/apiv1"
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	"google.golang.org/api/option"
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
	ctx := context.Background()

	projectID, ok := os.LookupEnv("GPC_PROJECT")
	if !ok {
		projectID = "testing-192515"
	}

	// Initialize the app with a custom auth variable, limiting the server's access
	ao := map[string]interface{}{
		"uid": "my-service-worker",
	}
	conf := &firebase.Config{
		ProjectID: projectID,
		StorageBucket: fmt.Sprintf("%s.appspot.com", projectID),
		AuthOverride: &ao,
	}

	// Fetch the service account key JSON file contents
	opt := option.WithCredentialsFile("firebase-adminsdk-testing-192515.json")

	// Initialize default app
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		log.Fatalf("firebase.NewApp: %v\n", err)
	}

	// Access firestore service from the default app
	firestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("app.Firestore: %v", err)
	}
}

func OnNovelCoverUpload(ctx context.Context, e GCSEvent) error {
	_, err := metadata.FromContext(ctx) // Change _ => meta
	if err != nil {
		log.Fatalf("metadata.FromContext: %v", err)
	}

	setNovelHasCover(ctx, e.Name[7:27])

	return nil
}

func setNovelHasCover(ctx context.Context, novelId string) {
	_, err := firestoreClient.Collection("novels").Doc(novelId).Set(ctx, map[string]interface{}{
		"cover": true,
	}, firestore.MergeAll)

	if err != nil {
		log.Fatalf("setNovelHasCover error: %v\n", err.Error())
	}
}