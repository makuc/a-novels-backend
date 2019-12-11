package main

import (
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"

	_ "image/gif"
	_ "image/png"
)

func main() {
	src, err := decode("pic2.png", "tmp.jpg")
	if err != nil {
		log.Fatalf("error converting: %v", err.Error())
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

	rThumb, rFull := PrepThumbAndFullRectangles(sX, sY)
	dstThumb := image.NewRGBA64(rThumb)
	dstFull := image.NewRGBA64(rFull)

	// Now scale images as Thumbnail and Full Size
	draw.BiLinear.Scale(dstThumb, rThumb, src, sr, draw.Src, nil)
	draw.BiLinear.Scale(dstFull, rFull, src, sr, draw.Src, nil)

	err = encode("z-thumb.jpg", dstThumb)
	err = encode("z-full.jpg", dstFull)
	if err != nil {
		return // nothing will change!
	}
}

// PrepThumbAndFullRectangles takes width and height to calculate optimal sizes
// for Thumbnail and Full image versions of the original
func PrepThumbAndFullRectangles(width int, height int) (image.Rectangle, image.Rectangle) {
	// Now figure out final sizes, since we don't want enlarged images...
	yThumb := 420
	xThumb := 280
	yFull := 2100
	xFull := 1400

	heightBasedOnWidth := CalcYBasedOnX(width)

	if heightBasedOnWidth <= height {
		// appropriate ratio for working with Width
		if xThumb < width {
			xThumb = width
			yThumb = CalcYBasedOnX(xThumb)
			xFull = xThumb
			yFull = yThumb
		} else if xFull < width {
			xFull = width
			yFull = CalcYBasedOnX(xFull)
		}
	} else {
		// appropriate ratio for working with Height
		if yThumb < height {
			yThumb = height
			xThumb = CalcXBasedOnY(yThumb)
			yFull = yThumb
			xFull = xThumb
		} else if yFull < height {
			yFull = height
			xFull = CalcXBasedOnY(yFull)
		}
	}

	rectThumb := image.Rect(0, 0, xThumb, yThumb)
	rectFull := image.Rect(0, 0, xFull, yFull)

	return rectThumb, rectFull
}

// CalcXBasedOnY calculates proportional Width based on provided Y (height)
func CalcXBasedOnY(y int) int {
	return y / 3 * 2
}

// CalcYBasedOnX calculates proportional Height based on provided X (width)
func CalcYBasedOnX(x int) int {
	return x / 2 * 3
}

func decode(srcName string, dstName string) (image.Image, error) {
	src, err := os.Open(srcName)
	defer src.Close()
	if err != nil {
		return nil, err
	}

	img, err := DecodeImage(src)
	if err != nil {
		return nil, err
	}
	return img, nil
}
func encode(dstName string, img image.Image) error {
	dst, err := os.Create(dstName)
	defer dst.Close()
	if err != nil {
		return err
	}

	return SaveJPG(img, dst)
}

// DecodeImage takes an input image of an unknown format and decodes it
func DecodeImage(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r) // encoded image, format name, error
	if err != nil {
		return nil, err
	}
	return img, nil
}

// SaveJPG convert the image to JPEG format and saves it to disk in JPEG of 85% quality
func SaveJPG(img image.Image, w io.Writer) error {
	opts := jpeg.Options{
		Quality: 85,
	}
	return jpeg.Encode(w, img, &opts)
}
