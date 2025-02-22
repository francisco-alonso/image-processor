package storage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

func DownloadImage(ctx context.Context, bucketName, fileName string) ([]byte, string, error){
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	obj := client.Bucket(bucketName).Object(fileName)

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object attributes: %w", err)
	}
	format := attrs.ContentType

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}

	return data, format, nil
}


func UploadImage(ctx context.Context, bucketName, fileName string, imageData []byte, format string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	obj := client.Bucket(bucketName).Object(fileName)
	writer := obj.NewWriter(ctx)
	writer.ContentType = format
	_, err = writer.Write(imageData)
	if err != nil {
		return fmt.Errorf("failed to write image data: %w", err)
	}

	return writer.Close()
}
