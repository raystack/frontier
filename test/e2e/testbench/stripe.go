package testbench

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/stripe/stripe-go/v79"

	"github.com/raystack/frontier/billing"
	"github.com/raystack/frontier/cmd"
	"github.com/raystack/frontier/config"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v79/client"

	"github.com/ory/dockertest/v3/docker"
	"github.com/raystack/salt/log"
)

const (
	stripeImage   = "stripe/stripe-mock"
	stripeVersion = "v0.193.0"
)

func StartStripeMock(logger log.Logger, network *docker.Network, pool *dockertest.Pool) (extPort string, close func() error, err error) {
	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   stripeImage,
		Tag:          stripeVersion,
		ExposedPorts: []string{"12111/tcp"}, // "12112/tcp"
		NetworkID:    network.ID,
	}, func(config *docker.HostConfig) {
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
		config.AutoRemove = true
	})
	if err != nil {
		return
	}

	if err = res.Expire(120); err != nil {
		return "", nil, err
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second

	extPort = res.GetPort("12111/tcp")
	stripeClientBuilder, _ := BuildStripeClient(extPort, "", recorder.ModePassthrough)
	cmd.GetStripeClientFunc = stripeClientBuilder
	stripeClient := stripeClientBuilder(logger, &config.Frontier{
		Billing: billing.Config{
			StripeKey: "sk_test_XXX",
		},
	})

	idemKey := uuid.New()
	if err = pool.Retry(func() error {
		customer, err := stripeClient.Customers.New(&stripe.CustomerParams{
			Params: stripe.Params{
				IdempotencyKey: stripe.String(idemKey.String()),
			},
			Email: stripe.String("test@testtest.com"),
			Name:  stripe.String("Test Customer"),
		})
		if customer.Email == "" {
			return fmt.Errorf("not ready")
		}
		return err
	}); err != nil {
		err = fmt.Errorf("could not connect to docker: %w", err)
		return
	}

	close = func() error {
		return res.Close()
	}
	return
}

func StartStripeRecorder(mode recorder.Mode) func() error {
	var closeFunc func() error
	cmd.GetStripeClientFunc, closeFunc = BuildStripeClient("", "", mode)
	return closeFunc
}

type StripeClientBuilder func(logger log.Logger, cfg *config.Frontier) *client.API

func BuildStripeClient(port, name string, mode recorder.Mode) (StripeClientBuilder, func() error) {
	var closer = func() error { return nil }
	var stripeURL *string
	if port != "" {
		stripeURL = stripe.String("http://localhost:" + port)
	}
	if name == "" {
		name = uuid.NewString()
	}
	if mode == -1 {
		mode = recorder.ModePassthrough
	}
	stripeHTTPClient := &http.Client{
		Timeout: 80 * time.Second,
	}

	if stripeURL == nil {
		r, err := recorder.NewWithOptions(&recorder.Options{
			CassetteName: fmt.Sprintf("testdata/cassettes/%s", name),
			Mode:         mode,
		})
		if err != nil {
			panic(err)
		}
		r.AddHook(func(i *cassette.Interaction) error {
			// remove all secret headers
			i.Request.Headers.Del("Authorization")
			i.Response.Headers.Del("Authorization")
			i.Request.Headers.Del("Cookie")
			i.Response.Headers.Del("Cookie")
			return nil
		}, recorder.BeforeSaveHook)
		stripeHTTPClient = r.GetDefaultClient()
		closer = func() error {
			return r.Stop() // Make sure recorder is stopped once done with it
		}
	}

	return func(logger log.Logger, cfg *config.Frontier) *client.API {
		stripeLogLevel := stripe.LevelError
		if cfg.Log.Level == "debug" {
			stripeLogLevel = stripe.LevelDebug
		}
		stripeBackends := &stripe.Backends{
			API: stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
				URL:        stripeURL,
				HTTPClient: stripeHTTPClient,
				LeveledLogger: &stripe.LeveledLogger{
					Level: stripeLogLevel,
				},
			}),
			Connect: stripe.GetBackend(stripe.ConnectBackend),
			Uploads: stripe.GetBackend(stripe.UploadsBackend),
		}
		stripeClient := client.New(cfg.Billing.StripeKey, stripeBackends)
		if cfg.Billing.StripeKey == "" {
			logger.Warn("stripe key is empty, billing services will be non-functional")
		} else {
			logger.Info("stripe client created")
		}
		return stripeClient
	}, closer
}
