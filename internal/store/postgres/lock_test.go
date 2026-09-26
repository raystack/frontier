package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/raystack/frontier/pkg/db"
)

type LockTestSuite struct {
	suite.Suite
	ctx    context.Context
	client *db.Client
}

func TestLocks(t *testing.T) {
	suite.Run(t, new(LockTestSuite))
}

func (s *LockTestSuite) SetupSuite() {
	var err error

	s.client, err = newTestClient()
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
}

func (s *LockTestSuite) SetupTest() {}

func (s *LockTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
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
