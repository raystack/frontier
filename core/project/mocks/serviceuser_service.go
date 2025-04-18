// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	serviceuser "github.com/raystack/frontier/core/serviceuser"
)

// ServiceuserService is an autogenerated mock type for the ServiceuserService type
type ServiceuserService struct {
	mock.Mock
}

type ServiceuserService_Expecter struct {
	mock *mock.Mock
}

func (_m *ServiceuserService) EXPECT() *ServiceuserService_Expecter {
	return &ServiceuserService_Expecter{mock: &_m.Mock}
}

// FilterSudos provides a mock function with given fields: ctx, ids
func (_m *ServiceuserService) FilterSudos(ctx context.Context, ids []string) ([]string, error) {
	ret := _m.Called(ctx, ids)

	if len(ret) == 0 {
		panic("no return value specified for FilterSudos")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([]string, error)); ok {
		return rf(ctx, ids)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) []string); ok {
		r0 = rf(ctx, ids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ServiceuserService_FilterSudos_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FilterSudos'
type ServiceuserService_FilterSudos_Call struct {
	*mock.Call
}

// FilterSudos is a helper method to define mock.On call
//   - ctx context.Context
//   - ids []string
func (_e *ServiceuserService_Expecter) FilterSudos(ctx interface{}, ids interface{}) *ServiceuserService_FilterSudos_Call {
	return &ServiceuserService_FilterSudos_Call{Call: _e.mock.On("FilterSudos", ctx, ids)}
}

func (_c *ServiceuserService_FilterSudos_Call) Run(run func(ctx context.Context, ids []string)) *ServiceuserService_FilterSudos_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]string))
	})
	return _c
}

func (_c *ServiceuserService_FilterSudos_Call) Return(_a0 []string, _a1 error) *ServiceuserService_FilterSudos_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ServiceuserService_FilterSudos_Call) RunAndReturn(run func(context.Context, []string) ([]string, error)) *ServiceuserService_FilterSudos_Call {
	_c.Call.Return(run)
	return _c
}

// GetByIDs provides a mock function with given fields: ctx, ids
func (_m *ServiceuserService) GetByIDs(ctx context.Context, ids []string) ([]serviceuser.ServiceUser, error) {
	ret := _m.Called(ctx, ids)

	if len(ret) == 0 {
		panic("no return value specified for GetByIDs")
	}

	var r0 []serviceuser.ServiceUser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([]serviceuser.ServiceUser, error)); ok {
		return rf(ctx, ids)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) []serviceuser.ServiceUser); ok {
		r0 = rf(ctx, ids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]serviceuser.ServiceUser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ServiceuserService_GetByIDs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByIDs'
type ServiceuserService_GetByIDs_Call struct {
	*mock.Call
}

// GetByIDs is a helper method to define mock.On call
//   - ctx context.Context
//   - ids []string
func (_e *ServiceuserService_Expecter) GetByIDs(ctx interface{}, ids interface{}) *ServiceuserService_GetByIDs_Call {
	return &ServiceuserService_GetByIDs_Call{Call: _e.mock.On("GetByIDs", ctx, ids)}
}

func (_c *ServiceuserService_GetByIDs_Call) Run(run func(ctx context.Context, ids []string)) *ServiceuserService_GetByIDs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]string))
	})
	return _c
}

func (_c *ServiceuserService_GetByIDs_Call) Return(_a0 []serviceuser.ServiceUser, _a1 error) *ServiceuserService_GetByIDs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ServiceuserService_GetByIDs_Call) RunAndReturn(run func(context.Context, []string) ([]serviceuser.ServiceUser, error)) *ServiceuserService_GetByIDs_Call {
	_c.Call.Return(run)
	return _c
}

// NewServiceuserService creates a new instance of ServiceuserService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewServiceuserService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ServiceuserService {
	mock := &ServiceuserService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
