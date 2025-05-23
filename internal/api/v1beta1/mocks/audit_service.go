// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	audit "github.com/raystack/frontier/core/audit"

	mock "github.com/stretchr/testify/mock"
)

// AuditService is an autogenerated mock type for the AuditService type
type AuditService struct {
	mock.Mock
}

type AuditService_Expecter struct {
	mock *mock.Mock
}

func (_m *AuditService) EXPECT() *AuditService_Expecter {
	return &AuditService_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, log
func (_m *AuditService) Create(ctx context.Context, log *audit.Log) error {
	ret := _m.Called(ctx, log)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *audit.Log) error); ok {
		r0 = rf(ctx, log)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AuditService_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type AuditService_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - log *audit.Log
func (_e *AuditService_Expecter) Create(ctx interface{}, log interface{}) *AuditService_Create_Call {
	return &AuditService_Create_Call{Call: _e.mock.On("Create", ctx, log)}
}

func (_c *AuditService_Create_Call) Run(run func(ctx context.Context, log *audit.Log)) *AuditService_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*audit.Log))
	})
	return _c
}

func (_c *AuditService_Create_Call) Return(_a0 error) *AuditService_Create_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AuditService_Create_Call) RunAndReturn(run func(context.Context, *audit.Log) error) *AuditService_Create_Call {
	_c.Call.Return(run)
	return _c
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *AuditService) GetByID(ctx context.Context, id string) (audit.Log, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetByID")
	}

	var r0 audit.Log
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (audit.Log, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) audit.Log); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(audit.Log)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AuditService_GetByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByID'
type AuditService_GetByID_Call struct {
	*mock.Call
}

// GetByID is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *AuditService_Expecter) GetByID(ctx interface{}, id interface{}) *AuditService_GetByID_Call {
	return &AuditService_GetByID_Call{Call: _e.mock.On("GetByID", ctx, id)}
}

func (_c *AuditService_GetByID_Call) Run(run func(ctx context.Context, id string)) *AuditService_GetByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *AuditService_GetByID_Call) Return(_a0 audit.Log, _a1 error) *AuditService_GetByID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AuditService_GetByID_Call) RunAndReturn(run func(context.Context, string) (audit.Log, error)) *AuditService_GetByID_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: ctx, filter
func (_m *AuditService) List(ctx context.Context, filter audit.Filter) ([]audit.Log, error) {
	ret := _m.Called(ctx, filter)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []audit.Log
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, audit.Filter) ([]audit.Log, error)); ok {
		return rf(ctx, filter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, audit.Filter) []audit.Log); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]audit.Log)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, audit.Filter) error); ok {
		r1 = rf(ctx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AuditService_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type AuditService_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
//   - filter audit.Filter
func (_e *AuditService_Expecter) List(ctx interface{}, filter interface{}) *AuditService_List_Call {
	return &AuditService_List_Call{Call: _e.mock.On("List", ctx, filter)}
}

func (_c *AuditService_List_Call) Run(run func(ctx context.Context, filter audit.Filter)) *AuditService_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(audit.Filter))
	})
	return _c
}

func (_c *AuditService_List_Call) Return(_a0 []audit.Log, _a1 error) *AuditService_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AuditService_List_Call) RunAndReturn(run func(context.Context, audit.Filter) ([]audit.Log, error)) *AuditService_List_Call {
	_c.Call.Return(run)
	return _c
}

// NewAuditService creates a new instance of AuditService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAuditService(t interface {
	mock.TestingT
	Cleanup(func())
}) *AuditService {
	mock := &AuditService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
