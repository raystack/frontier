package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/odpf/shield/middleware/basic_auth"
	"github.com/odpf/shield/middleware/casbin_authz"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/middleware/prefix"
	"github.com/odpf/shield/middleware/rulematch"
	"github.com/odpf/shield/store"
	"github.com/pkg/errors"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/memblob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"
)

// buildPipeline builds middleware sequence
func buildPipeline(logger log.Logger, proxy http.Handler, ruleRepo store.RuleRepository) http.Handler {
	// Note: execution order is bottom up
	prefixWare := prefix.New(logger, proxy)
	casbinAuthz := casbin_authz.New(logger, prefixWare)
	basicAuthn := basic_auth.New(logger, casbinAuthz)
	matchWare := rulematch.New(logger, basicAuthn, rulematch.NewRegexMatcher(ruleRepo))
	return matchWare
}

type blobFactory struct{}

func (o *blobFactory) New(ctx context.Context, storagePath, storageSecret string) (store.Bucket, error) {
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
		if storageSecret == "" {
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
		if prefix == "" {
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
