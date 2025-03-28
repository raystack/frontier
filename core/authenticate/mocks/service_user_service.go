// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	serviceuser "github.com/raystack/frontier/core/serviceuser"
	mock "github.com/stretchr/testify/mock"
)

// ServiceUserService is an autogenerated mock type for the ServiceUserService type
type ServiceUserService struct {
	mock.Mock
}

type ServiceUserService_Expecter struct {
	mock *mock.Mock
}

func (_m *ServiceUserService) EXPECT() *ServiceUserService_Expecter {
	return &ServiceUserService_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: ctx, id
func (_m *ServiceUserService) Get(ctx context.Context, id string) (serviceuser.ServiceUser, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 serviceuser.ServiceUser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (serviceuser.ServiceUser, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) serviceuser.ServiceUser); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(serviceuser.ServiceUser)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ServiceUserService_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type ServiceUserService_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *ServiceUserService_Expecter) Get(ctx interface{}, id interface{}) *ServiceUserService_Get_Call {
	return &ServiceUserService_Get_Call{Call: _e.mock.On("Get", ctx, id)}
}

func (_c *ServiceUserService_Get_Call) Run(run func(ctx context.Context, id string)) *ServiceUserService_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *ServiceUserService_Get_Call) Return(_a0 serviceuser.ServiceUser, _a1 error) *ServiceUserService_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ServiceUserService_Get_Call) RunAndReturn(run func(context.Context, string) (serviceuser.ServiceUser, error)) *ServiceUserService_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetByJWT provides a mock function with given fields: ctx, token
func (_m *ServiceUserService) GetByJWT(ctx context.Context, token string) (serviceuser.ServiceUser, error) {
	ret := _m.Called(ctx, token)

	if len(ret) == 0 {
		panic("no return value specified for GetByJWT")
	}

	var r0 serviceuser.ServiceUser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (serviceuser.ServiceUser, error)); ok {
		return rf(ctx, token)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) serviceuser.ServiceUser); ok {
		r0 = rf(ctx, token)
	} else {
		r0 = ret.Get(0).(serviceuser.ServiceUser)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ServiceUserService_GetByJWT_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByJWT'
type ServiceUserService_GetByJWT_Call struct {
	*mock.Call
}

// GetByJWT is a helper method to define mock.On call
//   - ctx context.Context
//   - token string
func (_e *ServiceUserService_Expecter) GetByJWT(ctx interface{}, token interface{}) *ServiceUserService_GetByJWT_Call {
	return &ServiceUserService_GetByJWT_Call{Call: _e.mock.On("GetByJWT", ctx, token)}
}

func (_c *ServiceUserService_GetByJWT_Call) Run(run func(ctx context.Context, token string)) *ServiceUserService_GetByJWT_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *ServiceUserService_GetByJWT_Call) Return(_a0 serviceuser.ServiceUser, _a1 error) *ServiceUserService_GetByJWT_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ServiceUserService_GetByJWT_Call) RunAndReturn(run func(context.Context, string) (serviceuser.ServiceUser, error)) *ServiceUserService_GetByJWT_Call {
	_c.Call.Return(run)
	return _c
}

// GetBySecret provides a mock function with given fields: ctx, clientID, clientSecret
func (_m *ServiceUserService) GetBySecret(ctx context.Context, clientID string, clientSecret string) (serviceuser.ServiceUser, error) {
	ret := _m.Called(ctx, clientID, clientSecret)

	if len(ret) == 0 {
		panic("no return value specified for GetBySecret")
	}

	var r0 serviceuser.ServiceUser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (serviceuser.ServiceUser, error)); ok {
		return rf(ctx, clientID, clientSecret)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) serviceuser.ServiceUser); ok {
		r0 = rf(ctx, clientID, clientSecret)
	} else {
		r0 = ret.Get(0).(serviceuser.ServiceUser)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, clientID, clientSecret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ServiceUserService_GetBySecret_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetBySecret'
type ServiceUserService_GetBySecret_Call struct {
	*mock.Call
}

// GetBySecret is a helper method to define mock.On call
//   - ctx context.Context
//   - clientID string
//   - clientSecret string
func (_e *ServiceUserService_Expecter) GetBySecret(ctx interface{}, clientID interface{}, clientSecret interface{}) *ServiceUserService_GetBySecret_Call {
	return &ServiceUserService_GetBySecret_Call{Call: _e.mock.On("GetBySecret", ctx, clientID, clientSecret)}
}

func (_c *ServiceUserService_GetBySecret_Call) Run(run func(ctx context.Context, clientID string, clientSecret string)) *ServiceUserService_GetBySecret_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *ServiceUserService_GetBySecret_Call) Return(_a0 serviceuser.ServiceUser, _a1 error) *ServiceUserService_GetBySecret_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ServiceUserService_GetBySecret_Call) RunAndReturn(run func(context.Context, string, string) (serviceuser.ServiceUser, error)) *ServiceUserService_GetBySecret_Call {
	_c.Call.Return(run)
	return _c
}

// NewServiceUserService creates a new instance of ServiceUserService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewServiceUserService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ServiceUserService {
	mock := &ServiceUserService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
