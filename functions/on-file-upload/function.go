package onfileupload

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	app             *firebase.App
	firestoreClient *firestore.Client
	bucket          *storage.BucketHandle

	storageClient *storage.Client
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
		log.Fatalf("missing ENV configuration for google cloud project: GCP_PROJECT")
	}

	// Initialize the app with a custom auth variable, limiting the server's access
	uid, ok := os.LookupEnv("worker_id")
	if !ok {
		log.Fatalf("set env variable `worker_id`")
	}
	// Initialize the app with a custom auth variable, limiting the server's access
	ao := map[string]interface{}{
		"uid": uid,
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

	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("storage.NewClient: %v", err)
	}

	if err != nil {
		log.Fatalf("vision.NewImageAnnotatorClient: %v", err)
	}
}

// OnFileUpload executes upon uploading a new file to the bucket.
// Executes all necessary validations, like processing novels' covers, users' profile pictures, etc.
func OnFileUpload(ctx context.Context, e GCSEvent) error {

	proceed, err := ExecuteWithLease(ctx)
	if err != nil {
		log.Printf("ExecuteWithLease: %v", err.Error())
		return err
	}
	if !proceed {
		return nil // this means this EventID was already completed
	}

	// Check if Covers
	log.Printf("path: %v", e.Name)

	// Cover successfully uploaded
	err = setNovelHasCover(ctx, e.Name[7:27])

	return nil
}

func setNovelHasCover(ctx context.Context, novelID string) error {
	_, err := firestoreClient.Collection("novels").Doc(novelID).Set(ctx, map[string]interface{}{
		"cover": true,
	}, firestore.MergeAll)

	if err != nil {
		log.Fatalf("firestoreClient.Doc: %v\n", err.Error())
		return err
	}
	return nil
}

func resizeImage(ctx context.Context, novelID string) {

}

// ExecuteWithLease tracks indempotence of the function based on the EventID specified in context metadata.
func ExecuteWithLease(ctx context.Context) (bool, error) {
	proceed := true

	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return false, err
	}

	ref := firestoreClient.Collection("events").Doc(meta.EventID)
	err = firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil && grpc.Code(err) != codes.NotFound {
			log.Printf("error getting ID doc: %v", err)
			return err
		}
		if doc.Exists() {
			if doc.Data()["done"].(bool) { // Since this Event was already processed, just exit
				proceed = false
				return nil
			}
			getLease, err := doc.DataAt("lease")
			if err != nil {
				return err
			}
			if val, ok := getLease.(time.Time); ok && time.Now().Before(val) {
				return errors.New("occupied, try later")
			}
		}
		leaseSecondsRaw, ok := os.LookupEnv("leaseSeconds")
		if !ok {
			leaseSecondsRaw = "60" // Default value, just in case
			log.Print("check env: leaseSeconds, using default: 60")
		}
		leaseSeconds, err := strconv.ParseInt(leaseSecondsRaw, 10, 32)
		if err != nil {
			return err
		}

		newValue := map[string]interface{}{
			"lease":     time.Now().Add(time.Second * time.Duration(leaseSeconds)),
			"updatedAt": time.Now(),
		}
		_, err = doc.DataAt("createdAt")
		if err != nil {
			if grpc.Code(err) == codes.NotFound {
				newValue["createdAt"] = time.Now()
			} else {
				log.Printf("can't get field `createdAt`: %v", err.Error())
				return err
			}
		}

		proceed = true
		return tx.Set(ref, newValue, firestore.MergeAll)
	})
	if err != nil {
		return false, err
	}
	return proceed, nil
}
