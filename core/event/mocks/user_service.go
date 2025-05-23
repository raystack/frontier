// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	user "github.com/raystack/frontier/core/user"
)

// UserService is an autogenerated mock type for the UserService type
type UserService struct {
	mock.Mock
}

type UserService_Expecter struct {
	mock *mock.Mock
}

func (_m *UserService) EXPECT() *UserService_Expecter {
	return &UserService_Expecter{mock: &_m.Mock}
}

// ListByOrg provides a mock function with given fields: ctx, orgID, roleFilter
func (_m *UserService) ListByOrg(ctx context.Context, orgID string, roleFilter string) ([]user.User, error) {
	ret := _m.Called(ctx, orgID, roleFilter)

	if len(ret) == 0 {
		panic("no return value specified for ListByOrg")
	}

	var r0 []user.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) ([]user.User, error)); ok {
		return rf(ctx, orgID, roleFilter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []user.User); ok {
		r0 = rf(ctx, orgID, roleFilter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]user.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, orgID, roleFilter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserService_ListByOrg_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListByOrg'
type UserService_ListByOrg_Call struct {
	*mock.Call
}

// ListByOrg is a helper method to define mock.On call
//   - ctx context.Context
//   - orgID string
//   - roleFilter string
func (_e *UserService_Expecter) ListByOrg(ctx interface{}, orgID interface{}, roleFilter interface{}) *UserService_ListByOrg_Call {
	return &UserService_ListByOrg_Call{Call: _e.mock.On("ListByOrg", ctx, orgID, roleFilter)}
}

func (_c *UserService_ListByOrg_Call) Run(run func(ctx context.Context, orgID string, roleFilter string)) *UserService_ListByOrg_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *UserService_ListByOrg_Call) Return(_a0 []user.User, _a1 error) *UserService_ListByOrg_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *UserService_ListByOrg_Call) RunAndReturn(run func(context.Context, string, string) ([]user.User, error)) *UserService_ListByOrg_Call {
	_c.Call.Return(run)
	return _c
}

// NewUserService creates a new instance of UserService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUserService(t interface {
	mock.TestingT
	Cleanup(func())
}) *UserService {
	mock := &UserService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
