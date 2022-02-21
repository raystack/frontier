package schema

import (
	"testing"

	"github.com/odpf/shield/internal/authz"
	authzmock "github.com/odpf/shield/internal/authz/mocks"
	"github.com/odpf/shield/internal/schema/mocks"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	service Service
	store   mocks.Store
	authz   authz.Authz
}

func (s *ServiceTestSuite) SetupTest() {
	s.store = mocks.Store{}
	s.authz = authz.Authz{
		Policy:     &authzmock.Policy{},
		Permission: &authzmock.Permission{},
	}
	s.service = Service{
		Store: &s.store,
		Authz: &s.authz,
	}
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
