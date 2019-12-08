package onuserupdate

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"context"
	"errors"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	client *firestore.Client
)

// FirestoreEvent is the payload of a Firestore event.
type FirestoreEvent struct {
	// OldValue   		UnknownValues	`json:"oldValue"`
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `firestore:"value"`
	UpdateMask FieldPaths     `firestore:"updateMask"`
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
	UID         StringValue `json:"uid"`
	DisplayName StringValue `json:"displayName"`
	Email       StringValue `json:"email"`
	//EmailVerified
	PhoneNumber StringValue    `json:"phoneNumber"`
	PhotoURL    StringValue    `json:"photoURL"`
	CreatedAt   TimestampValue `json:"createdAt"`
}

// UnknownValues is a helper type for figuring out structure of parsed data (by logging it)
type UnknownValues struct {
	Fields map[string]interface{} `json:"fields"`
}

// FieldPaths is a type for parsing `fieldPaths` data from Firestore Events
type FieldPaths struct {
	FieldPaths []string `json:"fieldPaths"`
}

// StringValue is a type for parsing `string` type from Firestore Events
type StringValue struct {
	StringValue string `json:"stringValue"`
}

// IntegerValue is a type for parsing `integer` type from Firestore Events
type IntegerValue struct {
	IntegerValue int64 `json:"integerValue"`
}

// TimestampValue is a type for parsing `Timestamp` type from Firestore Events
type TimestampValue struct {
	TimestampValue time.Time `json:"timestampValue"`
}

// BooleanValue is a type for parsing `boolean` type from Firestore Events
type BooleanValue struct {
	BooleanValue bool `json:"booleanValue"`
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

	proceed, err := ExecuteWithLease(ctx)
	if err != nil {
		log.Printf("ExecuteWithLease: %v", err.Error())
		return err
	}
	if !proceed {
		return nil // this means this EventID was already completed
	}

	// Continue with execution
	if e.OldValue.Fields.DisplayName.StringValue != e.Value.Fields.DisplayName.StringValue {
		// Display name has been changed! Now do the correct adjustments!
		err := changeNovelAuthor(ctx, e.Value.Fields.UID.StringValue, e.Value.Fields.DisplayName.StringValue)
		if err != nil {
			log.Printf("changeNovelsAuthor: %v", err.Error())
			return err
		}

		err = changeReviewsAuthor(ctx, e.Value.Fields.UID.StringValue, e.Value.Fields.DisplayName.StringValue)
		if err != nil {
			log.Printf("changeReviewsAuthor: %v", err.Error())
			return err
		}
	}

	return ExecuteMarkComplete(ctx)
}

func changeNovelAuthor(ctx context.Context, uid string, name string) error {
	progressRef, err := GetExecuteProgressRef(ctx)
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
	progressRef, err := GetExecuteProgressRef(ctx)
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

// ExecuteWithLease tracks indempotence of the function based on the EventID specified in context metadata.
func ExecuteWithLease(ctx context.Context) (bool, error) {
	proceed := true

	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return false, err
	}

	ref := client.Collection("user-events").Doc(meta.EventID)
	err = client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
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

// ExecuteMarkCompleteTransaction marks indempotent function as complete based
// on EventID specified in context metadata
func ExecuteMarkCompleteTransaction(ctx context.Context, tx *firestore.Transaction) error {
	ref, err := GetExecuteProgressRef(ctx)
	if err != nil {
		return err
	}
	return tx.Set(ref, map[string]interface{}{
		"done":      true,
		"updatedAt": time.Now(),
	}, firestore.MergeAll)
}

// ExecuteMarkCompleteBatch marks indempotent function as complete based
// on EventID specified in context metadata
func ExecuteMarkCompleteBatch(ctx context.Context, bx *firestore.WriteBatch) error {
	ref, err := GetExecuteProgressRef(ctx)
	if err != nil {
		return err
	}
	bx.Set(ref, map[string]interface{}{
		"done":      true,
		"updatedAt": time.Now(),
	}, firestore.MergeAll)

	return nil
}

// ExecuteMarkComplete marks indempotent function as complete based
// on EventID specified in context metadata
func ExecuteMarkComplete(ctx context.Context) error {
	ref, err := GetExecuteProgressRef(ctx)
	if err != nil {
		return err
	}
	_, err = ref.Set(ctx, map[string]interface{}{
		"done":      true,
		"updatedAt": time.Now(),
	}, firestore.MergeAll)

	return err
}

// GetExecuteProgressRef returns a document reference for the document tracking progress of
// indempotent completion of this function based on EventID specified in context metadata
func GetExecuteProgressRef(ctx context.Context) (*firestore.DocumentRef, error) {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	return client.Collection("user-events").Doc(meta.EventID), nil
}
