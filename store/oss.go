package store

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
)

type OSSStorage struct {
	bucket *oss.Bucket
	prefix string
}

func NewOSSStorage(bucket *oss.Bucket, prefix string) *OSSStorage {
	return &OSSStorage{bucket: bucket, prefix: prefix}
}

func (s *OSSStorage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	var files []string
	marker := ""
	for {
		lsRes, err := s.bucket.ListObjects(oss.Marker(marker), oss.Prefix(prefix))
		if err != nil {
			return nil, err
		}

		for _, obj := range lsRes.Objects {
			files = append(files, obj.Key)
		}

		if !lsRes.IsTruncated {
			break
		}
		marker = lsRes.NextMarker
	}
	if files == nil {
		return []string{}, nil
	}
	return files, nil
}

func (s *OSSStorage) ReadFile(ctx context.Context, key string) ([]byte, error) {
	body, err := s.bucket.GetObject(key)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *OSSStorage) WriteFile(ctx context.Context, key string, data io.Reader) error {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(data)
	if err != nil {
		return fmt.Errorf("error reading data before writing to OSS: %w", err)
	}
	return s.bucket.PutObject(key, buf)
}

func (s *OSSStorage) Close() error {
	return nil
}
