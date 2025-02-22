package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockPubSubClient struct {
	shouldFail bool
}

func (m *MockPubSubClient) ReceiveMessage(ctx context.Context, msgData []byte) error {
	var obj GCSObject
	if err := json.Unmarshal(msgData, &obj); err != nil {
		return errors.New("failed to parse message")
	}

	if m.shouldFail {
		return errors.New("failed to process message")
	}

	return nil
}


func TestPubSubProcessing(t *testing.T) {
	validMsg := GCSObject{Bucket: "source-bucket", Name: "image.jpg"}
	invalidMsg := []byte("invalid json")

	validData, _ := json.Marshal(validMsg)

	tests := []struct {
		name        string
		data        []byte
		mockFailure bool
		expectError bool
	}{
		{"Valid PubSub Message", validData, false, false},
		{"Invalid PubSub Message", invalidMsg, false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &MockPubSubClient{shouldFail: tc.mockFailure}
			err := mockClient.ReceiveMessage(context.Background(), tc.data)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
