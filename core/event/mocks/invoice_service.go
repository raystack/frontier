// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// InvoiceService is an autogenerated mock type for the InvoiceService type
type InvoiceService struct {
	mock.Mock
}

type InvoiceService_Expecter struct {
	mock *mock.Mock
}

func (_m *InvoiceService) EXPECT() *InvoiceService_Expecter {
	return &InvoiceService_Expecter{mock: &_m.Mock}
}

// TriggerSyncByProviderID provides a mock function with given fields: ctx, id
func (_m *InvoiceService) TriggerSyncByProviderID(ctx context.Context, id string) error {
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

// InvoiceService_TriggerSyncByProviderID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TriggerSyncByProviderID'
type InvoiceService_TriggerSyncByProviderID_Call struct {
	*mock.Call
}

// TriggerSyncByProviderID is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *InvoiceService_Expecter) TriggerSyncByProviderID(ctx interface{}, id interface{}) *InvoiceService_TriggerSyncByProviderID_Call {
	return &InvoiceService_TriggerSyncByProviderID_Call{Call: _e.mock.On("TriggerSyncByProviderID", ctx, id)}
}

func (_c *InvoiceService_TriggerSyncByProviderID_Call) Run(run func(ctx context.Context, id string)) *InvoiceService_TriggerSyncByProviderID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *InvoiceService_TriggerSyncByProviderID_Call) Return(_a0 error) *InvoiceService_TriggerSyncByProviderID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *InvoiceService_TriggerSyncByProviderID_Call) RunAndReturn(run func(context.Context, string) error) *InvoiceService_TriggerSyncByProviderID_Call {
	_c.Call.Return(run)
	return _c
}

// NewInvoiceService creates a new instance of InvoiceService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewInvoiceService(t interface {
	mock.TestingT
	Cleanup(func())
}) *InvoiceService {
	mock := &InvoiceService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
