// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	policy "github.com/raystack/frontier/core/policy"
	mock "github.com/stretchr/testify/mock"
)

// PolicyService is an autogenerated mock type for the PolicyService type
type PolicyService struct {
	mock.Mock
}

type PolicyService_Expecter struct {
	mock *mock.Mock
}

func (_m *PolicyService) EXPECT() *PolicyService_Expecter {
	return &PolicyService_Expecter{mock: &_m.Mock}
}

// List provides a mock function with given fields: ctx, f
func (_m *PolicyService) List(ctx context.Context, f policy.Filter) ([]policy.Policy, error) {
	ret := _m.Called(ctx, f)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []policy.Policy
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, policy.Filter) ([]policy.Policy, error)); ok {
		return rf(ctx, f)
	}
	if rf, ok := ret.Get(0).(func(context.Context, policy.Filter) []policy.Policy); ok {
		r0 = rf(ctx, f)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]policy.Policy)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, policy.Filter) error); ok {
		r1 = rf(ctx, f)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PolicyService_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type PolicyService_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
//   - f policy.Filter
func (_e *PolicyService_Expecter) List(ctx interface{}, f interface{}) *PolicyService_List_Call {
	return &PolicyService_List_Call{Call: _e.mock.On("List", ctx, f)}
}

func (_c *PolicyService_List_Call) Run(run func(ctx context.Context, f policy.Filter)) *PolicyService_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(policy.Filter))
	})
	return _c
}

func (_c *PolicyService_List_Call) Return(_a0 []policy.Policy, _a1 error) *PolicyService_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PolicyService_List_Call) RunAndReturn(run func(context.Context, policy.Filter) ([]policy.Policy, error)) *PolicyService_List_Call {
	_c.Call.Return(run)
	return _c
}

// NewPolicyService creates a new instance of PolicyService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPolicyService(t interface {
	mock.TestingT
	Cleanup(func())
}) *PolicyService {
	mock := &PolicyService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
