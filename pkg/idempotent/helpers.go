package idempotent

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
)

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
	return Client.Collection("user-events").Doc(meta.EventID), nil
}
