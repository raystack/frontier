package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/raystack/frontier/core/webhook"
	"github.com/raystack/frontier/pkg/metadata"

	"github.com/google/uuid"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/suite"
)

type WebhookEndpointRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.WebhookEndpointRepository
}

func (s *WebhookEndpointRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewWebhookEndpointRepository(s.client, []byte("kmm4ECoWU21K2ZoyTcYLd6w7DfhoUoap"))
}

func (s *WebhookEndpointRepositoryTestSuite) SetupTest() {
}

func (s *WebhookEndpointRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *WebhookEndpointRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *WebhookEndpointRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_WEBHOOK_ENDPOINTS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *WebhookEndpointRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description string
		ToCreate    webhook.Endpoint
		Expected    webhook.Endpoint
		ErrString   string
	}

	var testCases = []testCase{
		{
			Description: "should create a webhook",
			ToCreate: webhook.Endpoint{
				ID:          uuid.NewString(),
				Description: "web1",
				URL:         "http://localhost:8080",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				SubscribedEvents: []string{"event1", "event2"},
				Secrets: []webhook.Secret{
					{
						ID:    "secret1",
						Value: "value1",
					},
				},
			},
			Expected: webhook.Endpoint{
				Description: "web1",
				URL:         "http://localhost:8080",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				SubscribedEvents: []string{"event1", "event2"},
				Secrets: []webhook.Secret{
					{
						ID:    "secret1",
						Value: "value1",
					},
				},
				State:    webhook.Enabled,
				Metadata: metadata.Metadata{},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.ToCreate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
			if diff := cmp.Diff(tc.Expected, got, cmpopts.IgnoreFields(webhook.Endpoint{}, "ID", "CreatedAt", "UpdatedAt")); diff != "" {
				s.T().Fatalf("expected -, got +:\n%s", diff)
			}
		})
	}
}

func (s *WebhookEndpointRepositoryTestSuite) TestList() {
	type testCase struct {
		Description string
		Expected    []webhook.Endpoint
		ErrString   string
	}

	created1, err := s.repository.Create(s.ctx, webhook.Endpoint{
		ID:          uuid.NewString(),
		Description: "web1",
		URL:         "http://localhost:8080",
		State:       webhook.Enabled,
	})
	s.Assert().NoError(err)
	created2, err := s.repository.Create(s.ctx, webhook.Endpoint{
		ID:          uuid.NewString(),
		Description: "web2",
		URL:         "http://localhost:8080",
		State:       webhook.Disabled,
	})
	s.Assert().NoError(err)

	var testCases = []testCase{
		{
			Description: "should get all webhooks",
			Expected: []webhook.Endpoint{
				created2,
				created1,
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.List(s.ctx, webhook.EndpointFilter{})
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
			if diff := cmp.Diff(tc.Expected, got, cmpopts.IgnoreFields(webhook.Endpoint{}, "ID", "CreatedAt", "UpdatedAt")); diff != "" {
				s.T().Fatalf("expected (-), got (+):\n%s", diff)
			}
		})
	}
}

func (s *WebhookEndpointRepositoryTestSuite) TestUpdateByID() {
	type testCase struct {
		Description string
		ToUpdate    webhook.Endpoint
		Expected    webhook.Endpoint
		ErrString   string
	}

	created1, err := s.repository.Create(s.ctx, webhook.Endpoint{
		ID:          uuid.NewString(),
		Description: "web1",
		URL:         "http://localhost:8080",
		State:       webhook.Enabled,
	})
	s.Assert().NoError(err)

	var testCases = []testCase{
		{
			Description: "should return error if webhook not found",
			ToUpdate: webhook.Endpoint{
				ID:          uuid.New().String(),
				Description: "not-exist",
			},
			ErrString: webhook.ErrNotFound.Error(),
		},
		{
			Description: "should return error if webhook id is empty",
			ErrString:   webhook.ErrInvalidDetail.Error(),
		},
		{
			Description: "should update webhook",
			ToUpdate: webhook.Endpoint{
				ID:          created1.ID,
				Description: "web1-updated",
				URL:         "http://localhost:8080",
			},
			Expected: webhook.Endpoint{
				ID:          created1.ID,
				Description: "web1-updated",
				URL:         "http://localhost:8080",
				State:       webhook.Enabled,
				Metadata:    metadata.Metadata{},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.UpdateByID(s.ctx, tc.ToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
			if diff := cmp.Diff(tc.Expected, got, cmpopts.IgnoreFields(webhook.Endpoint{}, "ID", "CreatedAt", "UpdatedAt")); diff != "" {
				s.T().Fatalf("expected -, got +:\n%s", diff)
			}
		})
	}
}

func TestWebhookEndpointRepository(t *testing.T) {
	suite.Run(t, new(WebhookEndpointRepositoryTestSuite))
}
