// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	subscription "github.com/raystack/frontier/billing/subscription"
)

// SubscriptionService is an autogenerated mock type for the SubscriptionService type
type SubscriptionService struct {
	mock.Mock
}

type SubscriptionService_Expecter struct {
	mock *mock.Mock
}

func (_m *SubscriptionService) EXPECT() *SubscriptionService_Expecter {
	return &SubscriptionService_Expecter{mock: &_m.Mock}
}

// List provides a mock function with given fields: ctx, filter
func (_m *SubscriptionService) List(ctx context.Context, filter subscription.Filter) ([]subscription.Subscription, error) {
	ret := _m.Called(ctx, filter)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []subscription.Subscription
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, subscription.Filter) ([]subscription.Subscription, error)); ok {
		return rf(ctx, filter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, subscription.Filter) []subscription.Subscription); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]subscription.Subscription)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, subscription.Filter) error); ok {
		r1 = rf(ctx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SubscriptionService_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type SubscriptionService_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
//   - filter subscription.Filter
func (_e *SubscriptionService_Expecter) List(ctx interface{}, filter interface{}) *SubscriptionService_List_Call {
	return &SubscriptionService_List_Call{Call: _e.mock.On("List", ctx, filter)}
}

func (_c *SubscriptionService_List_Call) Run(run func(ctx context.Context, filter subscription.Filter)) *SubscriptionService_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(subscription.Filter))
	})
	return _c
}

func (_c *SubscriptionService_List_Call) Return(_a0 []subscription.Subscription, _a1 error) *SubscriptionService_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SubscriptionService_List_Call) RunAndReturn(run func(context.Context, subscription.Filter) ([]subscription.Subscription, error)) *SubscriptionService_List_Call {
	_c.Call.Return(run)
	return _c
}

// NewSubscriptionService creates a new instance of SubscriptionService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSubscriptionService(t interface {
	mock.TestingT
	Cleanup(func())
}) *SubscriptionService {
	mock := &SubscriptionService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
