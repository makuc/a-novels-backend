package transaction

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/makuc/a-novels-backend/pkg/idempotent"
)

// ExecuteMarkCompleteTransaction marks indempotent function as complete based
// on EventID specified in context metadata
func ExecuteMarkCompleteTransaction(ctx context.Context, tx *firestore.Transaction) error {
	ref, err := idempotent.GetExecuteProgressRef(ctx)
	if err != nil {
		return err
	}
	return tx.Set(ref, map[string]interface{}{
		"done":      true,
		"updatedAt": time.Now(),
	}, firestore.MergeAll)
}
