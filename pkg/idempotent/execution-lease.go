package idempotent

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	// Client variable needs to be injected a pointer to an initialized Firestore Client
	Client *firestore.Client
)

// ExecuteWithLease tracks indempotence of the function based on the EventID specified in context metadata.
func ExecuteWithLease(ctx context.Context, leaseSeconds int) (bool, error) {
	proceed := true

	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return false, err
	}

	ref := Client.Collection("user-events").Doc(meta.EventID)
	err = Client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil && grpc.Code(err) != codes.NotFound {
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

		newValue := map[string]interface{}{
			"lease":     time.Now().Add(time.Second * time.Duration(leaseSeconds)),
			"updatedAt": time.Now(),
		}

		_, err = doc.DataAt("createdAt")
		if err != nil {
			if grpc.Code(err) == codes.NotFound {
				newValue["createdAt"] = time.Now()
			} else {
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
