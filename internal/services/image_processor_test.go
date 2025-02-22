package services

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to generate a test image
func createTestImage(format string) ([]byte, string, error) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Fill image with a solid color
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
		}
	}

	buffer := new(bytes.Buffer)
	switch format {
	case "image/jpeg":
		err := jpeg.Encode(buffer, img, nil)
		return buffer.Bytes(), "image/jpeg", err
	case "image/png":
		err := png.Encode(buffer, img)
		return buffer.Bytes(), "image/png", err
	default:
		return nil, "", fmt.Errorf("unsupported format")
	}
}

func TestProcessImage(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		expectError bool
	}{
		{"Valid JPEG Image", "image/jpeg", false},
		{"Valid PNG Image", "image/png", false},
		{"Unsupported Format", "image/bmp", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var imgData []byte
			var format string
			var err error
			
			if tc.expectError {
				// Provide invalid image data
				imgData = []byte("this is not an image")
				format = tc.format
			} else {
				imgData, format, err = createTestImage(tc.format)
				if err != nil {
					t.Fatalf("failed to create test image: %v", err)
				}
			}

			resizedData, err := ProcessImage(imgData, format)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, resizedData)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resizedData)
				assert.Greater(t, len(resizedData), 0, "Resized image should have data")
			}
		})
	}
}
