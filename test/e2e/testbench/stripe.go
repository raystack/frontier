package testbench

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/client"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/raystack/salt/log"
)

const (
	stripeImage   = "stripe/stripe-mock"
	stripeVersion = "latest"
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

	// Configure a backend for stripe-mock and set it for both the API and
	// Uploads (unlike the real Stripe API, stripe-mock supports both these
	// backends).
	stripeMockBackend := stripe.GetBackendWithConfig(
		stripe.APIBackend,
		&stripe.BackendConfig{
			URL:           stripe.String("http://localhost:" + extPort),
			HTTPClient:    http.DefaultClient,
			LeveledLogger: stripe.DefaultLeveledLogger,
		},
	)
	stripe.SetBackend(stripe.APIBackend, stripeMockBackend)
	stripe.SetBackend(stripe.UploadsBackend, stripeMockBackend)
	stripeClient := client.New("sk_test_myTestKey", nil)
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
