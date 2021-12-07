package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
	"github.com/pkg/profile"
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
	proxyTermChan         = make(chan os.Signal, 1)
	ruleCacheRefreshDelay = time.Minute * 2
)

// h2c reverse proxy
func proxyCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	c := &cli.Command{
		Use:     "proxy",
		Short:   "Start proxy over h2c",
		Example: "shield serve proxy",
		RunE: func(c *cli.Command, args []string) error {
			if profiling := os.Getenv("SHIELD_PROFILE"); profiling == "true" || profiling == "1" {
				defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
			}

			baseCtx, baseCancel := context.WithCancel(context.Background())
			defer baseCancel()

			var cleanUpFunc []func() error
			var cleanUpProxies []func(ctx context.Context) error
			for _, service := range appConfig.Proxy.Services {
				if service.ResourcesConfigPath == "" {
					return errors.New("ruleset field cannot be left empty")
				}
				resoucesConfigFS, err := (&blobFactory{}).New(baseCtx, service.RulesPath, service.RulesPathSecret)
				if err != nil {
					return err
				}

				resoucesRepo := blobstore.NewResourcesRepository(logger, resoucesConfigFS)

				if err := resoucesRepo.InitCache(baseCtx, ruleCacheRefreshDelay); err != nil {
					return err
				}

				if service.RulesPath == "" {
					return errors.New("ruleset field cannot be left empty")
				}
				ruleFS, err := (&blobFactory{}).New(baseCtx, service.RulesPath, service.RulesPathSecret)
				if err != nil {
					return err
				}

				// TODO: option to use default http round tripper for http1.1 backends
				h2cProxy := proxy.NewH2c(proxy.NewH2cRoundTripper(logger), proxy.NewDirector())

				ruleRepo := blobstore.NewRuleRepository(logger, ruleFS)
				if err := ruleRepo.InitCache(baseCtx, ruleCacheRefreshDelay); err != nil {
					return err
				}
				cleanUpFunc = append(cleanUpFunc, ruleRepo.Close)
				pipeline := buildPipeline(logger, h2cProxy, ruleRepo)
				go func(thisService config.Service, handler http.Handler) {
					proxyURL := fmt.Sprintf("%s:%d", thisService.Host, thisService.Port)
					logger.Info("starting h2c proxy", "url", proxyURL)

					mux := http.NewServeMux()
					mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprintf(w, "pong")
					})
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
					if err := proxySrv.Serve(proxyListener); err != nil && err != http.ErrServerClosed {
						logger.Fatal("failed to serve", "err", err)
					}
					cleanUpProxies = append(cleanUpProxies, proxySrv.Shutdown)
				}(service, pipeline)
			}

			// we'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
			signal.Notify(proxyTermChan, os.Interrupt, os.Kill, syscall.SIGTERM)

			// block until we receive our signal
			<-proxyTermChan
			for _, f := range cleanUpFunc {
				if err := f(); err != nil {
					logger.Warn("error occurred during shutdown", "err", err)
				}
			}
			for _, f := range cleanUpProxies {
				shutdownCtx, shutdownCancel := context.WithTimeout(baseCtx, time.Second*20)
				if err := f(shutdownCtx); err != nil {
					shutdownCancel()
					logger.Warn("error occurred during shutdown", "err", err)
					continue
				}
				shutdownCancel()
			}
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
