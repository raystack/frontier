package postgres_test

import (
	"context"
	"testing"

	"github.com/ory/dockertest"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/suite"

	"github.com/raystack/frontier/pkg/db"
)

type LockTestSuite struct {
	suite.Suite
	ctx      context.Context
	client   *db.Client
	pool     *dockertest.Pool
	resource *dockertest.Resource
}

func TestLocks(t *testing.T) {
	suite.Run(t, new(LockTestSuite))
}

func (s *LockTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
}

func (s *LockTestSuite) SetupTest() {}

func (s *LockTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *LockTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *LockTestSuite) cleanup() error {
	queries := []string{}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *LockTestSuite) TestLocking() {
	ctx := context.Background()
	s.Run("lock and unlock over an id successfully", func() {
		id := "test-id"
		lock, err := s.client.TryLock(ctx, id)
		s.Require().NoError(err)

		err = lock.Unlock(ctx)
		s.Require().NoError(err)
	})
	s.Run("shouldn't allow same lock more then once", func() {
		id := "test-id"
		lock, err := s.client.TryLock(ctx, id)
		s.Require().NoError(err)

		_, err = s.client.TryLock(ctx, id)
		s.Require().ErrorIs(err, db.ErrLockBusy)

		err = lock.Unlock(ctx)
		s.Require().NoError(err)
	})
}
