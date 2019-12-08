package batch

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/makuc/a-novels-backend/pkg/idempotent"
)

// ExecuteMarkCompleteBatch marks indempotent function as complete based
// on EventID specified in context metadata
func ExecuteMarkCompleteBatch(ctx context.Context, bx *firestore.WriteBatch) error {
	ref, err := idempotent.GetExecuteProgressRef(ctx)
	if err != nil {
		return err
	}
	bx.Set(ref, map[string]interface{}{
		"done":      true,
		"updatedAt": time.Now(),
	}, firestore.MergeAll)

	return nil
}
