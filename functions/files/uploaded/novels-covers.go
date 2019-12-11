package uploaded

import (
	"context"
	"fmt"
	"image"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/makuc/a-novels-backend/pkg/gcp/gcse"
	"golang.org/x/image/draw"
)

func processNovelsCovers(ctx context.Context, e gcse.GCSEvent) error {
	nameLen := len(e.Name)
	if nameLen < 32 || e.Name[:6] != "novels" || e.Name[nameLen-4:nameLen] != "orig" {
		return nil // Execution not target at Covers
	}

	if coverProcessedRaw, ok := progress["cover"]; ok {
		if coverProcessed, ok := coverProcessedRaw.(bool); ok && coverProcessed {
			return nil
		}
	}

	novelID := e.Name[7:27]
	objSrc := bucket.Object(e.Name)
	var src image.Image

	err := func() error {
		// Prepare reader for Original Picture
		rc, err := objSrc.NewReader(ctx)
		defer rc.Close()
		if err != nil {
			return err
		}

		src, err = decode(rc)
		if err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		return err // Maybe Bucket is BUSY, retry
	}

	sX := src.Bounds().Dx() // width
	sY := src.Bounds().Dy() // height
	sYBasedOnX := sX / 2 * 3
	sXBasedOnY := sY / 3 * 2

	var sr image.Rectangle
	if sYBasedOnX <= sY {
		// Use sYBasedOnX
		yOffset := (sY - sYBasedOnX) / 2
		sr = image.Rect(0, yOffset, sX, sYBasedOnX)
	} else {
		// Use sXBasedOnY
		xOffset := (sX - sXBasedOnY) / 2
		sr = image.Rect(xOffset, 0, sXBasedOnY, sY)
	}

	nameThumb := fmt.Sprintf("novels/%v/thumb.jpg", novelID)
	nameFull := fmt.Sprintf("novels/%v/full.jpg", novelID)

	rThumb, rFull := prepThumbAndFullRectangles(sX, sY)

	// Prepare THUMB
	err = func() error {
		dst := image.NewRGBA64(rThumb)
		draw.NearestNeighbor.Scale(dst, rThumb, src, sr, draw.Src, nil)
		obj := bucket.Object(nameThumb).NewWriter(ctx)
		defer obj.Close()
		if err = encode(obj, dst); err != nil {
			return err // Maybe Bucket is BUSY, retry
		}

		return nil
	}()
	if err != nil {
		return err
	}

	// Prepare FULL
	err = func() error {
		dst := image.NewRGBA64(rFull)
		draw.NearestNeighbor.Scale(dst, rFull, src, sr, draw.Src, nil)
		obj := bucket.Object(nameFull).NewWriter(ctx)
		defer obj.Close()
		if err = encode(obj, dst); err != nil {
			return err // Maybe Bucket is BUSY, retry
		}
		return nil
	}()
	if err != nil {
		return err
	}

	if err = setNovelHasCover(ctx, novelID); err != nil {
		return err // Maybe Firebase is BUSY, retry
	}

	// Since we processed the cover, delete orig, to save space
	if err = objSrc.Delete(ctx); err != nil {
		return err
	}

	// Succeeded, now mark as such
	progressRef.Set(ctx, map[string]interface{}{
		"cover": true,
	}, firestore.MergeAll)

	return nil
}

func setNovelHasCover(ctx context.Context, novelID string) error {

	aclThumb := bucket.Object(fmt.Sprintf("novels/%v/thumb.jpg", novelID)).ACL()
	if err := aclThumb.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return err
	}
	aclFull := bucket.Object(fmt.Sprintf("novels/%v/full.jpg", novelID)).ACL()
	if err := aclFull.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return err
	}

	// PLAYING AROUND
	nameThumb := fmt.Sprintf("novels/%v/thumb.jpg", novelID)
	nameFull := fmt.Sprintf("novels/%v/full.jpg", novelID)
	attrsThumb, err := bucket.Object(nameThumb).Attrs(ctx)
	if err != nil {
		return err
	}
	attrsFull, err := bucket.Object(nameFull).Attrs(ctx)
	if err != nil {
		return err
	}

	_, err = firestoreClient.Collection("novels").Doc(novelID).Set(ctx, map[string]interface{}{
		"coverThumbURL": attrsThumb.MediaLink,
		"coverFullURL":  attrsFull.MediaLink,
	}, firestore.MergeAll)

	if err != nil {
		return err
	}
	return nil
}
