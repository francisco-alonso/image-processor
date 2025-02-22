package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock Storage Client
type MockStorageClient struct {
	shouldFail bool
}

func (m *MockStorageClient) DownloadImage(ctx context.Context, bucketName, fileName string) ([]byte, string, error) {
	if m.shouldFail {
		return nil, "", errors.New("failed to read object")
	}
	return []byte("fake image data"), "image/jpeg", nil
}

func TestDownloadImage(t *testing.T) {
	tests := []struct {
		name        string
		mockFailure bool
		expectError bool
	}{
		{"Valid Image", false, false},
		{"Bucket Not Found", true, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockStorageClient{shouldFail: tc.mockFailure}
			_, _, err := mockClient.DownloadImage(context.Background(), "test-bucket", "test.jpg")

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
