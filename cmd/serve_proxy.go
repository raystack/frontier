package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/odpf/shield/middleware/basic_auth"
	"github.com/odpf/shield/middleware/casbin_authz"

	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/middleware/prefix"
	"github.com/odpf/shield/middleware/rulematch"
	"github.com/odpf/shield/proxy"
	"github.com/odpf/shield/store"
	blobstore "github.com/odpf/shield/store/blob"
	"github.com/pkg/errors"
	cli "github.com/spf13/cobra"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/blob/memblob"
	"gocloud.dev/gcp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/oauth2/google"
)

var (
	proxyTermChan = make(chan os.Signal, 1)
)

// h2c reverse proxy
func proxyCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	c := &cli.Command{
		Use:     "proxy",
		Short:   "Start proxy over h2c",
		Example: "shield serve proxy",
		RunE: func(c *cli.Command, args []string) error {
			ctx := context.Background()
			for _, service := range appConfig.Proxy.Services {
				blobFS, err := (&blobFactory{}).New(ctx, service.RulesPath, "")
				if err != nil {
					return err
				}

				h2cProxy := proxy.NewH2c(proxy.NewH2cRoundTripper(), proxy.NewDirector())
				ruleRepo := blobstore.NewRuleRepository(blobFS)
				pipeline := buildPipeline(logger, h2cProxy, ruleRepo)

				go func(thisService config.Service, handler http.Handler) {
					proxyURL := fmt.Sprintf("%s:%d", thisService.Host, thisService.Port)
					fmt.Println("starting h2cProxy at", proxyURL)

					mux := http.NewServeMux()
					mux.Handle("/", handler)

					//create a tcp listener
					proxyListener, err := net.Listen("tcp", proxyURL)
					if err != nil {
						logger.Fatal("failed to listen", "err", err)
					}

					proxySrv := http.Server{
						Addr:    proxyURL,
						Handler: h2c.NewHandler(mux, &http2.Server{}),
					}
					if err := proxySrv.Serve(proxyListener); err != nil {
						logger.Fatal("failed to serve", "err", err)
					}
				}(service, pipeline)
			}

			// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
			signal.Notify(proxyTermChan, os.Interrupt)
			signal.Notify(proxyTermChan, os.Kill)
			signal.Notify(proxyTermChan, syscall.SIGTERM)

			// Block until we receive our signal.
			<-proxyTermChan

			//TODO: shutdown all proxies

			return nil
		},
	}
	return c
}

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
	parsedURL, err := url.Parse(storagePath)
	if err != nil {
		return nil, err
	}

	switch parsedURL.Scheme {
	case "gs":
		if storageSecret == "" {
			return nil, errors.Errorf("%s secret not configured for fs", storagePath)
		}
		creds, err := google.CredentialsFromJSON(ctx, []byte(storageSecret), "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return nil, err
		}
		client, err := gcp.NewHTTPClient(
			gcp.DefaultTransport(),
			gcp.CredentialsTokenSource(creds))
		if err != nil {
			return nil, err
		}

		gcsBucket, err := gcsblob.OpenBucket(ctx, client, parsedURL.Host, nil)
		if err != nil {
			return nil, err
		}
		// create a *blob.Bucket
		prefix := fmt.Sprintf("%s/", strings.Trim(parsedURL.Path, "/\\"))
		return blob.PrefixedBucket(gcsBucket, prefix), nil
	case "file":
		return fileblob.OpenBucket(parsedURL.Path, &fileblob.Options{
			CreateDir: true,
			Metadata:  fileblob.MetadataDontWrite,
		})
	case "mem":
		return memblob.OpenBucket(nil), nil
	}
	return nil, errors.Errorf("unsupported storage config %s", storagePath)
}
