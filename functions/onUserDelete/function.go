package onUserCreate

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"log"
	"os"
)

var client *firestore.Client

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

type AuthEvent struct {
	UID string `json:"uid"`
}

func OnUserDelete(ctx context.Context, e AuthEvent) error {
	// Remove the obsolete user document
	_, err := client.Collection("users").Doc(e.UID).Delete(ctx)
	if err != nil {
		log.Fatalf("users.Delete: %v", err)
	}

	return nil
}
