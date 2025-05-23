// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	rql "github.com/raystack/salt/rql"
	mock "github.com/stretchr/testify/mock"

	userorgs "github.com/raystack/frontier/core/aggregates/userorgs"
)

// UserOrgsService is an autogenerated mock type for the UserOrgsService type
type UserOrgsService struct {
	mock.Mock
}

type UserOrgsService_Expecter struct {
	mock *mock.Mock
}

func (_m *UserOrgsService) EXPECT() *UserOrgsService_Expecter {
	return &UserOrgsService_Expecter{mock: &_m.Mock}
}

// Search provides a mock function with given fields: ctx, id, query
func (_m *UserOrgsService) Search(ctx context.Context, id string, query *rql.Query) (userorgs.UserOrgs, error) {
	ret := _m.Called(ctx, id, query)

	if len(ret) == 0 {
		panic("no return value specified for Search")
	}

	var r0 userorgs.UserOrgs
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *rql.Query) (userorgs.UserOrgs, error)); ok {
		return rf(ctx, id, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *rql.Query) userorgs.UserOrgs); ok {
		r0 = rf(ctx, id, query)
	} else {
		r0 = ret.Get(0).(userorgs.UserOrgs)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *rql.Query) error); ok {
		r1 = rf(ctx, id, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserOrgsService_Search_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Search'
type UserOrgsService_Search_Call struct {
	*mock.Call
}

// Search is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - query *rql.Query
func (_e *UserOrgsService_Expecter) Search(ctx interface{}, id interface{}, query interface{}) *UserOrgsService_Search_Call {
	return &UserOrgsService_Search_Call{Call: _e.mock.On("Search", ctx, id, query)}
}

func (_c *UserOrgsService_Search_Call) Run(run func(ctx context.Context, id string, query *rql.Query)) *UserOrgsService_Search_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*rql.Query))
	})
	return _c
}

func (_c *UserOrgsService_Search_Call) Return(_a0 userorgs.UserOrgs, _a1 error) *UserOrgsService_Search_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *UserOrgsService_Search_Call) RunAndReturn(run func(context.Context, string, *rql.Query) (userorgs.UserOrgs, error)) *UserOrgsService_Search_Call {
	_c.Call.Return(run)
	return _c
}

// NewUserOrgsService creates a new instance of UserOrgsService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUserOrgsService(t interface {
	mock.TestingT
	Cleanup(func())
}) *UserOrgsService {
	mock := &UserOrgsService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
