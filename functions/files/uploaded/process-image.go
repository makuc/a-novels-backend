package uploaded

import (
	"image"
	_ "image/gif" // Needed for image conversion from GIF
	"image/jpeg"
	_ "image/png" // Needed for image conversion from PNG
	"io"
)

// PrepThumbAndFullRectangles takes width and height to calculate optimal sizes
// for Thumbnail and Full image versions of the original
func prepThumbAndFullRectangles(width int, height int) (image.Rectangle, image.Rectangle) {
	// Now figure out final sizes, since we don't want enlarged images...
	yThumb := 420
	xThumb := 280
	yFull := 2100
	xFull := 1400

	heightBasedOnWidth := calcYBasedOnX(width)

	if heightBasedOnWidth <= height {
		// appropriate ratio for working with Width
		if xThumb > width {
			xThumb = width
			yThumb = calcYBasedOnX(xThumb)
			xFull = xThumb
			yFull = yThumb
		} else if xFull > width {
			xFull = width
			yFull = calcYBasedOnX(xFull)
		}
	} else {
		// appropriate ratio for working with Height
		if yThumb > height {
			yThumb = height
			xThumb = calcXBasedOnY(yThumb)
			yFull = yThumb
			xFull = xThumb
		} else if yFull > height {
			yFull = height
			xFull = calcXBasedOnY(yFull)
		}
	}

	rectThumb := image.Rect(0, 0, xThumb, yThumb)
	rectFull := image.Rect(0, 0, xFull, yFull)

	return rectThumb, rectFull
}

// CalcXBasedOnY calculates proportional Width based on provided Y (height)
func calcXBasedOnY(y int) int {
	return y / 3 * 2
}

// CalcYBasedOnX calculates proportional Height based on provided X (width)
func calcYBasedOnX(x int) int {
	return x / 2 * 3
}

func decode(src io.Reader) (image.Image, error) {
	img, err := decodeImage(src)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func encode(dst io.Writer, img image.Image) error {
	return saveJPG(img, dst)
}

// DecodeImage takes an input image of an unknown format and decodes it
func decodeImage(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r) // encoded image, format name, error
	if err != nil {
		return nil, err
	}
	return img, nil
}

// SaveJPG convert the image to JPEG format and saves it to disk in JPEG of 85% quality
func saveJPG(img image.Image, w io.Writer) error {
	opts := jpeg.Options{
		Quality: 85,
	}
	return jpeg.Encode(w, img, &opts)
}
