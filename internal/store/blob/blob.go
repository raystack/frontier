package blob

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/memblob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"
)

type Bucket interface {
	WriteAll(ctx context.Context, key string, p []byte, opts *blob.WriterOptions) error
	ReadAll(ctx context.Context, key string) ([]byte, error)
	List(opts *blob.ListOptions) *blob.ListIterator
	Delete(ctx context.Context, key string) error
	Close() error
}

func NewStore(ctx context.Context, storagePath, storageSecret string) (Bucket, error) {
	if strings.TrimSpace(storagePath) == "" {
		return memblob.OpenBucket(nil), nil
	}
	var errBadSecretURL = errors.Errorf(`unsupported storage config %s, possible schemes supported: "env:// file:// val://" for example: "val://username:password"`, storageSecret)
	var errBadStorageURL = errors.Errorf("unsupported storage config %s", storagePath)

	var storageSecretValue []byte
	if storageSecret != "" {
		parsedSecretURL, err := url.Parse(storageSecret)
		if err != nil {
			return nil, errBadSecretURL
		}
		switch parsedSecretURL.Scheme {
		case "env":
			{
				storageSecretValue = []byte(os.Getenv(parsedSecretURL.Hostname()))
			}
		case "file":
			{
				fileContent, err := ioutil.ReadFile(parsedSecretURL.Path)
				if err != nil {
					return nil, errors.Wrap(err, "failed to read secret content at "+parsedSecretURL.Path)
				}
				storageSecretValue = fileContent
			}
		case "val":
			{
				storageSecretValue = []byte(parsedSecretURL.Hostname())
			}
		default:
			return nil, errBadSecretURL
		}
	}

	parsedStorageURL, err := url.Parse(storagePath)
	if err != nil {
		return nil, errors.Wrap(err, errBadStorageURL.Error())
	}
	switch parsedStorageURL.Scheme {
	case "gs":
		if strings.TrimSpace(storageSecret) == "" {
			return nil, errors.Errorf("%s secret not configured for fs", storagePath)
		}
		creds, err := google.CredentialsFromJSON(ctx, storageSecretValue, "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return nil, err
		}
		client, err := gcp.NewHTTPClient(
			gcp.DefaultTransport(),
			gcp.CredentialsTokenSource(creds))
		if err != nil {
			return nil, err
		}

		gcsBucket, err := gcsblob.OpenBucket(ctx, client, parsedStorageURL.Host, nil)
		if err != nil {
			return nil, err
		}
		// create a *blob.Bucket
		prefix := fmt.Sprintf("%s/", strings.Trim(parsedStorageURL.Path, "/\\"))
		if strings.TrimSpace(prefix) == "" {
			return gcsBucket, nil
		}
		return blob.PrefixedBucket(gcsBucket, prefix), nil
	case "file":
		return fileblob.OpenBucket(parsedStorageURL.Path, &fileblob.Options{
			CreateDir: true,
			Metadata:  fileblob.MetadataDontWrite,
		})
	case "mem":
		return memblob.OpenBucket(nil), nil
	}
	return nil, errBadStorageURL
}
