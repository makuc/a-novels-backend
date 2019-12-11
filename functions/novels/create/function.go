package create

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"context"
	firebase "firebase.google.com/go"
	"log"
	"os"
	"time"
)

var client *firestore.Client

// FirestoreEvent is the payload of a Firestore event.
type FirestoreEvent struct {
	OldValue   FirestoreValue `firestore:"oldValue"`
	Value      FirestoreValue `firestore:"value"`
	UpdateMask struct {
		FieldPaths []string `firestore:"fieldPaths"`
	} `firestore:"updateMask"`
}

// FirestoreValue holds Firestore fields.
type FirestoreValue struct {
	CreateTime time.Time `firestore:"createTime"`
	// Fields is the data for this value. The type depends on the format of your
	// database. Log the interface{} value and inspect the result to see a JSON
	// representation of your database fields.
	Fields     NovelPart `firestore:"fields"`
	Name       string    `firestore:"name"`
	UpdateTime time.Time `firestore:"updateTime"`
}
type NovelPart struct {
	ID struct {
		StringValue string `firestore:"stringValue"`
	} `firestore:"id"`
	Title struct {
		StringValue string `firestore:"stringValue"`
	} `firestore:"title.stringValue"`
}
type Novel struct {
	Id     string `firestore:"id"`
	Title  string `firestore:"title"`
	Author Author `firestore:"author"`
	//editors

	CoverURL  string `firestore:"coverURL""`
	Published bool   `firestore:"published"`

	CreatedAt time.Time `firestore:"createdAt.timestampValue"`
	UpdatedAt time.Time `firestore:"updatedAt.timestampValue"`

	Description string   `firestore:"description"`
	Genres      []Genre  `firestore:"genres"`
	Tags        []string `firestore:"tags"`

	NFavorites int64 `firestore:"nFavorites"`

	NRatings    int64 `firestore:"nRatings"`
	StoryRating int64 `firestore:"storyRating"`
	StyleRating int64 `firestore:"styleRating"`
	CharsRating int64 `firestore:"charsRating"`
	WorldRating int64 `firestore:"worldRating"`
	GrammRating int64 `firestore:"grammRating"`
}
type Author struct {
	Uid         string `firestore:"uid"`
	DisplayName string `firestore:"displayName"`
}
type Genre struct {
	Name        string `firestore:"name"`
	Description string `firestore:"description"`
}

func init() {
	ctx := context.Background()

	projectID, ok := os.LookupEnv("GPC_PROJECT")
	if !ok {
		projectID = "testing-192515"
	}

	conf := &firebase.Config{
		ProjectID: projectID,
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

// OnNovelCreate executes when a document in `novels` collection is created
func OnNovelCreate(ctx context.Context, e FirestoreEvent) error {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		log.Fatalf("metadata.FromContext: %v", err)
	}
	log.Printf("Function triggered by change to: %v", meta.Resource)
	log.Printf("ID direct: %v", e.Value.Fields.Id)
	log.Printf("ID child: %v", e.Value.Fields.Id.StringValue)
	log.Printf("Title direct: %s", e.Value.Fields.Title)
	log.Printf("Title child: %s", e.Value.Fields.Title.StringValue)

	return nil
}
