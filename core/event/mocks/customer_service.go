// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	customer "github.com/raystack/frontier/billing/customer"

	mock "github.com/stretchr/testify/mock"
)

// CustomerService is an autogenerated mock type for the CustomerService type
type CustomerService struct {
	mock.Mock
}

type CustomerService_Expecter struct {
	mock *mock.Mock
}

func (_m *CustomerService) EXPECT() *CustomerService_Expecter {
	return &CustomerService_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, _a1, offline
func (_m *CustomerService) Create(ctx context.Context, _a1 customer.Customer, offline bool) (customer.Customer, error) {
	ret := _m.Called(ctx, _a1, offline)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 customer.Customer
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, customer.Customer, bool) (customer.Customer, error)); ok {
		return rf(ctx, _a1, offline)
	}
	if rf, ok := ret.Get(0).(func(context.Context, customer.Customer, bool) customer.Customer); ok {
		r0 = rf(ctx, _a1, offline)
	} else {
		r0 = ret.Get(0).(customer.Customer)
	}

	if rf, ok := ret.Get(1).(func(context.Context, customer.Customer, bool) error); ok {
		r1 = rf(ctx, _a1, offline)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CustomerService_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type CustomerService_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 customer.Customer
//   - offline bool
func (_e *CustomerService_Expecter) Create(ctx interface{}, _a1 interface{}, offline interface{}) *CustomerService_Create_Call {
	return &CustomerService_Create_Call{Call: _e.mock.On("Create", ctx, _a1, offline)}
}

func (_c *CustomerService_Create_Call) Run(run func(ctx context.Context, _a1 customer.Customer, offline bool)) *CustomerService_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(customer.Customer), args[2].(bool))
	})
	return _c
}

func (_c *CustomerService_Create_Call) Return(_a0 customer.Customer, _a1 error) *CustomerService_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CustomerService_Create_Call) RunAndReturn(run func(context.Context, customer.Customer, bool) (customer.Customer, error)) *CustomerService_Create_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: ctx, flt
func (_m *CustomerService) List(ctx context.Context, flt customer.Filter) ([]customer.Customer, error) {
	ret := _m.Called(ctx, flt)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []customer.Customer
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, customer.Filter) ([]customer.Customer, error)); ok {
		return rf(ctx, flt)
	}
	if rf, ok := ret.Get(0).(func(context.Context, customer.Filter) []customer.Customer); ok {
		r0 = rf(ctx, flt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]customer.Customer)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, customer.Filter) error); ok {
		r1 = rf(ctx, flt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CustomerService_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type CustomerService_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
//   - flt customer.Filter
func (_e *CustomerService_Expecter) List(ctx interface{}, flt interface{}) *CustomerService_List_Call {
	return &CustomerService_List_Call{Call: _e.mock.On("List", ctx, flt)}
}

func (_c *CustomerService_List_Call) Run(run func(ctx context.Context, flt customer.Filter)) *CustomerService_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(customer.Filter))
	})
	return _c
}

func (_c *CustomerService_List_Call) Return(_a0 []customer.Customer, _a1 error) *CustomerService_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CustomerService_List_Call) RunAndReturn(run func(context.Context, customer.Filter) ([]customer.Customer, error)) *CustomerService_List_Call {
	_c.Call.Return(run)
	return _c
}

// TriggerSyncByProviderID provides a mock function with given fields: ctx, id
func (_m *CustomerService) TriggerSyncByProviderID(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for TriggerSyncByProviderID")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CustomerService_TriggerSyncByProviderID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TriggerSyncByProviderID'
type CustomerService_TriggerSyncByProviderID_Call struct {
	*mock.Call
}

// TriggerSyncByProviderID is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *CustomerService_Expecter) TriggerSyncByProviderID(ctx interface{}, id interface{}) *CustomerService_TriggerSyncByProviderID_Call {
	return &CustomerService_TriggerSyncByProviderID_Call{Call: _e.mock.On("TriggerSyncByProviderID", ctx, id)}
}

func (_c *CustomerService_TriggerSyncByProviderID_Call) Run(run func(ctx context.Context, id string)) *CustomerService_TriggerSyncByProviderID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CustomerService_TriggerSyncByProviderID_Call) Return(_a0 error) *CustomerService_TriggerSyncByProviderID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CustomerService_TriggerSyncByProviderID_Call) RunAndReturn(run func(context.Context, string) error) *CustomerService_TriggerSyncByProviderID_Call {
	_c.Call.Return(run)
	return _c
}

// NewCustomerService creates a new instance of CustomerService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCustomerService(t interface {
	mock.TestingT
	Cleanup(func())
}) *CustomerService {
	mock := &CustomerService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
