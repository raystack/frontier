package store

import (
	"context"
	"fmt"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob" // Import for GCS support
	"io"
	"io/ioutil"
)

type GCSStorage struct {
	bucket *blob.Bucket
	prefix string
}

func NewGCSStorage(bucket *blob.Bucket, prefix string) *GCSStorage {
	return &GCSStorage{bucket: bucket, prefix: prefix}
}

func (s *GCSStorage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	var files []string
	iter := s.bucket.List(&blob.ListOptions{Prefix: prefix})
	for {
		obj, err := iter.Next(ctx)
		if err == io.EOF { // Just break instead of returning an error
			break
		}
		if err != nil {
			return nil, err
		}
		files = append(files, obj.Key)
	}
	return files, nil
}

func (s *GCSStorage) ReadFile(ctx context.Context, key string) ([]byte, error) {
	return s.bucket.ReadAll(ctx, key)
}

func (s *GCSStorage) WriteFile(ctx context.Context, key string, data io.Reader) error {
	dataByte, err := ioutil.ReadAll(data)
	if err != nil {
		fmt.Println("error writing in a file!", err.Error())
		return fmt.Errorf("error reading data before writing to file: %w", err)
	}
	return s.bucket.WriteAll(ctx, key, dataByte, nil)
}

func (s *GCSStorage) Close() error {
	return s.bucket.Close()
}
