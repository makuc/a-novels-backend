package uploaded

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"github.com/makuc/a-novels-backend/pkg/gcp/gcse"
	"github.com/makuc/a-novels-backend/pkg/idempotent"
)

var (
	app             *firebase.App
	firestoreClient *firestore.Client
	bucket          *storage.BucketHandle

	// tmp variables
	progress    map[string]interface{}
	progressRef *firestore.DocumentRef
	projectID   string
	bucketName  string
)

func init() {
	ctx := context.Background()

	projectID, ok := os.LookupEnv("GPC_PROJECT")
	if !ok {
		projectID = "testing-192515"
	}

	// Initialize the app with a custom auth variable, limiting the server's access
	uid, ok := os.LookupEnv("WorkerID")
	if !ok {
		log.Fatal("Must config ENV variable: WorkerID")
	}

	// Initialize the app with a custom auth variable, limiting the server's access
	ao := map[string]interface{}{
		"uid": uid,
	}
	bucketName = fmt.Sprintf("%v.appspot.com", projectID)
	conf := &firebase.Config{
		ProjectID:     projectID,
		StorageBucket: bucketName,
		AuthOverride:  &ao,
	}

	// Fetch the service account key JSON file contents
	//opt := option.WithCredentialsFile("adminsdk-credentials.json")

	// Initialize default app
	app, err := firebase.NewApp(ctx, conf /*, opt*/)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err.Error())
	}

	// oauth2.TokenSource

	// Access firestore service from the default app
	firestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("error preparing firebase client: %v", err.Error())
	}

	idempotent.Client = firestoreClient

	// Access storage services from the default app
	storageClient, err := app.Storage(ctx)
	if err != nil {
		log.Fatalf("error preparing storage client: %v", err.Error())
	}

	bucket, err = storageClient.DefaultBucket()
	if err != nil {
		log.Fatalf("error loading default bucket: %v", err.Error())
	}

}

// OnFileUploaded executes when a new file is uploaded to the storage bucket.
// Performs all the necessary file transformations (for images), etc., as needed.
func OnFileUploaded(ctx context.Context, e gcse.GCSEvent) error {
	//return testiranje(ctx, e)

	if proceed, err := idempotent.ExecuteWithLease(ctx); err != nil || !proceed {
		return err
	}

	// Prepare OBJ with PROGRESS Data
	progressRef, err := idempotent.GetExecuteProgressRef(ctx)
	if err != nil {
		return err
	}
	data, err := progressRef.Get(ctx)
	if err != nil {
		return err
	}
	progress = data.Data()

	if err := processNovelsCovers(ctx, e); err != nil {
		// something went wrong while processing novel's cover, retry
		return err
	}

	idempotent.ExecuteMarkComplete(ctx)
	return nil

}
