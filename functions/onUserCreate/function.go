package onUserCreate

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"log"
	"os"
	"time"
)

var client *firestore.Client

type AuthEvent struct {
	UID           string `json:"uid"`
	DisplayName   string `json:"displayName"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	PhoneNumber   string `json:"phoneNumber"`
	PhotoURL      string `json:"photoURL"`

	Metadata struct {
		CreatedAt      time.Time `json:"createdAt"`
		CreationTime   time.Time `json:"creationTime"`
		LastSignInTime time.Time `json:"lastSignInTime"`
	} `json:"metadata"`
}
type UserProfile struct {
	UID           string `firestore:"uid"`
	DisplayName   string `firestore:"displayName"`
	Email         string `firestore:"email,omitempty"`
	EmailVerified bool   `firestore:"emailVerified,omitempty"`
	PhoneNumber   string `firestore:"phoneNumber,omitempty"`
	PhotoURL      string `firestore:"photoURL,omitempty"`

	CreatedAt time.Time `firestore:"createdAt"`
	Birthday  time.Time `firestore:"birthday,omitempty"`
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

func OnUserCreate(ctx context.Context, e AuthEvent) error {
	// Create the new user document
	_, err := client.Collection("users").Doc(e.UID).Set(ctx, UserProfile{
		UID:           e.UID,
		DisplayName:   e.DisplayName,
		Email:         e.Email,
		EmailVerified: e.EmailVerified,
		PhoneNumber:   e.PhoneNumber,
		PhotoURL:      e.PhotoURL,

		CreatedAt: e.Metadata.CreatedAt,
	})
	if err != nil {
		log.Fatalf("users.Add: %v", err)
	}

	return nil
}
