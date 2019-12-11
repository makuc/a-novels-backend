package deleted

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"github.com/makuc/a-novels-backend/pkg/idempotent"
)

var (
	app             *firebase.App
	firestoreClient *firestore.Client
	bucket          *storage.BucketHandle
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
	app, err := firebase.NewApp(ctx, conf /*, opt*/)
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

	idempotent.Client = firestoreClient
}

// OnFileDeleted executes when a file is deleted from the storage bucket.
// Performs all the necessary corrections (in DB) if necessary.
func OnFileDeleted(ctx context.Context, e GCSEvent) error {

	log.Printf("Resource state: %v", e.ResourceState) // == not-found ?? Shorter for deleted objects...

	proceed, err := idempotent.ExecuteWithLease(ctx)
	if err != nil {
		return err
	}
	if !proceed {
		return nil // EventID was already processed
	}

	_, err = metadata.FromContext(ctx) // Change _ => meta
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
