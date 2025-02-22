package services

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/nfnt/resize"
)

func ProcessImage(data []byte, format string) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	resizedImg := resize.Resize(300, 0, img, resize.Lanczos3)

	var buf bytes.Buffer
	switch format {
	case "image/jpeg":
		err = jpeg.Encode(&buf, resizedImg, nil)
	case "image/png":
		err = png.Encode(&buf, resizedImg)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %w", err)
	}

	return buf.Bytes(), nil
}
