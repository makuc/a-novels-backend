package update

import (
	"context"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/makuc/a-novels-backend/pkg/gcp"
	"github.com/makuc/a-novels-backend/pkg/idempotent"
	"google.golang.org/api/iterator"
)

var (
	client *firestore.Client
)

// FirestoreEvent is the payload of a Firestore event.
type FirestoreEvent struct {
	// OldValue   		UnknownValues	`json:"oldValue"`
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `firestore:"value"`
	UpdateMask gcp.FieldPaths `firestore:"updateMask"`
}

// FirestoreValue holds Firestore fields.
type FirestoreValue struct {
	CreateTime time.Time `json:"createTime"`
	// Fields is the data for this value.
	Fields     UserProfile `json:"fields"`
	Name       string      `json:"name"`
	UpdateTime time.Time   `json:"updateTime"`
}

// UserProfile is a struct containing field for UserProfile
type UserProfile struct {
	UID         gcp.StringValue `json:"uid"`
	DisplayName gcp.StringValue `json:"displayName"`
	Email       gcp.StringValue `json:"email"`
	//EmailVerified
	PhoneNumber gcp.StringValue    `json:"phoneNumber"`
	PhotoURL    gcp.StringValue    `json:"photoURL"`
	CreatedAt   gcp.TimestampValue `json:"createdAt"`
}

func init() {
	ctx := context.Background()

	projectID, ok := os.LookupEnv("GPC_PROJECT")
	if !ok {
		log.Fatalf("missing ENV configuration for google cloud project: GCP_PROJECT")
	}

	// Initialize the app with a custom auth variable, limiting the server's access
	uid, ok := os.LookupEnv("worker_id")
	if !ok {
		log.Fatalf("set env variable `worker_id`")
	}
	ao := map[string]interface{}{
		"uid": uid,
	}
	conf := &firebase.Config{
		ProjectID:    projectID,
		AuthOverride: &ao,
	}
	// Initialize default app
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalf("firebase.NewApp: %v\n", err)
	}
	// Access firestore service from the default app
	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("app.Firestore: %v", err)
	}
}

// OnUserUpdate executes when relevant entry in User collection is UPDATED
func OnUserUpdate(ctx context.Context, e FirestoreEvent) error {
	// log.Printf("Function triggered by change to: %v", meta.Resource)

	proceed, err := idempotent.ExecuteWithLease(ctx)
	if err != nil {
		log.Printf("ExecuteWithLease: %v", err.Error())
		return err
	}
	if !proceed {
		return nil // this means this EventID was already completed
	}

	// Continue with execution
	if e.OldValue.Fields.DisplayName.Value != e.Value.Fields.DisplayName.Value {
		// Display name has been changed! Now do the correct adjustments!
		err := changeNovelAuthor(ctx, e.Value.Fields.UID.Value, e.Value.Fields.DisplayName.Value)
		if err != nil {
			log.Printf("changeNovelsAuthor: %v", err.Error())
			return err
		}

		err = changeReviewsAuthor(ctx, e.Value.Fields.UID.Value, e.Value.Fields.DisplayName.Value)
		if err != nil {
			log.Printf("changeReviewsAuthor: %v", err.Error())
			return err
		}
	}

	return idempotent.ExecuteMarkComplete(ctx)
}

func changeNovelAuthor(ctx context.Context, uid string, name string) error {
	progressRef, err := idempotent.GetExecuteProgressRef(ctx)
	if err != nil {
		return err
	}

	doc, err := progressRef.Get(ctx)
	if err != nil {
		return err
	}

	// Fetch time the event was executed, to only affect older docs
	rawData := doc.Data()
	novelsDone := false
	novelsDoneRaw, ok := rawData["novelsDone"]
	if ok {
		novelsDone, ok = novelsDoneRaw.(bool)
		if !ok {
			log.Print("error asserting type of `novelsDone`")
		}
		if novelsDone {
			return nil // we are done here
		}
	}

	createdRaw, ok := rawData["createdAt"]
	if !ok {
		log.Printf("`createdAt` doesn't exist")
		return nil // No use retrying, result won't change
	}
	created, ok := createdRaw.(time.Time)
	if !ok {
		log.Printf("error assering type of `createdAt`")
		return nil // No use retrying here, result won't change
	}

	// Fetch what was already processed
	novelsStep := 0
	novelsStepRaw, ok := rawData["novelsStep"]
	if ok {
		novelsStep, ok = novelsStepRaw.(int)
		if !ok {
			log.Print("error asserting type of `novelsStep`")
			return nil // No use retrying here, result won't change
		}
	}

	// Determine size of each step (since batch write MAX is: 500)
	stepSize := 400

	for {
		queryNovels := client.Collection("novels").
			Where("author.uid", "==", uid).
			OrderBy("createdAt", firestore.Desc).
			Where("createdAt", "<", created).
			Offset(novelsStep * stepSize).
			Limit(stepSize)

		// Now we actually execute the rewrite
		batch := client.Batch()
		docsAffected := 0

		iter := queryNovels.Documents(ctx)
		for {
			doc, err := iter.Next()
			if err != nil {
				if err == iterator.Done {
					break // end of this loop
				}
				log.Printf("error iterating documents: %v", err.Error())
			}
			updatedData := map[string]interface{}{
				"author": map[string]interface{}{
					"uid":         uid,
					"displayName": name,
				},
			}
			batch.Set(doc.Ref, updatedData, firestore.MergeAll)
			docsAffected++
		}
		iter.Stop()

		// All docs (for this step) processed, save state
		progressNewValue := map[string]interface{}{
			"updatedAt":  time.Now(),
			"novelsStep": novelsStep + 1,
		}

		if docsAffected == 0 {
			// We are done here
			progressNewValue["novelsDone"] = true
		}

		batch.Set(progressRef, progressNewValue, firestore.MergeAll)

		_, err := batch.Commit(ctx)
		return err
	}
}

func changeReviewsAuthor(ctx context.Context, uid string, name string) error {
	progressRef, err := idempotent.GetExecuteProgressRef(ctx)
	if err != nil {
		return err
	}

	doc, err := progressRef.Get(ctx)
	if err != nil {
		return err
	}

	// Fetch time the event was executed, to only affect older docs
	rawData := doc.Data()
	reviewsDone := false
	reviewsDoneRaw, ok := rawData["reviewsDone"]
	if ok {
		reviewsDone, ok = reviewsDoneRaw.(bool)
		if !ok {
			log.Print("error asserting type of `reviewsDone`")
		}
		if reviewsDone {
			return nil // we are done here
		}
	}

	createdRaw, ok := rawData["createdAt"]
	if !ok {
		log.Printf("`createdAt` doesn't exist")
		return nil // No use retrying, result won't change
	}
	created, ok := createdRaw.(time.Time)
	if !ok {
		log.Printf("error assering type of `createdAt`")
		return nil // No use retrying here, result won't change
	}

	// Fetch what was already processed
	reviewsStep := 0
	reviewsStepRaw, ok := rawData["reviewsStep"]
	if ok {
		reviewsStep, ok = reviewsStepRaw.(int)
		if !ok {
			log.Print("error asserting type of `reviewsStep`")
			return nil // No use retrying here, result won't change
		}
	}

	// Determine size of each step (since batch write MAX is: 500)
	stepSize := 400

	for {
		query := client.CollectionGroup("reviews").
			Where("author.uid", "==", uid).
			OrderBy("createdAt", firestore.Desc).
			Where("createdAt", "<", created).
			Offset(reviewsStep * stepSize).
			Limit(stepSize)

		// Now we actually execute the rewrite
		batch := client.Batch()
		docsAffected := 0

		iter := query.Documents(ctx)
		for {
			doc, err := iter.Next()
			if err != nil {
				if err == iterator.Done {
					break // end of this loop
				}
				log.Printf("error iterating documents: %v", err.Error())
			}
			updatedData := map[string]interface{}{
				"author": map[string]interface{}{
					"uid":         uid,
					"displayName": name,
				},
			}
			batch.Set(doc.Ref, updatedData, firestore.MergeAll)
			docsAffected++
		}
		iter.Stop()

		// All docs (for this step) processed, save state
		progressNewValue := map[string]interface{}{
			"updatedAt":   time.Now(),
			"reviewsStep": reviewsStep + 1,
		}

		if docsAffected == 0 {
			// We are done here
			progressNewValue["reviewsDone"] = true
		}

		batch.Set(progressRef, progressNewValue, firestore.MergeAll)

		_, err := batch.Commit(ctx)
		return err
	}
}
