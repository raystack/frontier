// Code generated by mockery v2.40.2. DO NOT EDIT.

package mocks

import (
	context "context"

	credit "github.com/raystack/frontier/billing/credit"
	mock "github.com/stretchr/testify/mock"
)

// CreditService is an autogenerated mock type for the CreditService type
type CreditService struct {
	mock.Mock
}

type CreditService_Expecter struct {
	mock *mock.Mock
}

func (_m *CreditService) EXPECT() *CreditService_Expecter {
	return &CreditService_Expecter{mock: &_m.Mock}
}

// GetBalance provides a mock function with given fields: ctx, accountID
func (_m *CreditService) GetBalance(ctx context.Context, accountID string) (int64, error) {
	ret := _m.Called(ctx, accountID)

	if len(ret) == 0 {
		panic("no return value specified for GetBalance")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (int64, error)); ok {
		return rf(ctx, accountID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) int64); ok {
		r0 = rf(ctx, accountID)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, accountID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreditService_GetBalance_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetBalance'
type CreditService_GetBalance_Call struct {
	*mock.Call
}

// GetBalance is a helper method to define mock.On call
//   - ctx context.Context
//   - accountID string
func (_e *CreditService_Expecter) GetBalance(ctx interface{}, accountID interface{}) *CreditService_GetBalance_Call {
	return &CreditService_GetBalance_Call{Call: _e.mock.On("GetBalance", ctx, accountID)}
}

func (_c *CreditService_GetBalance_Call) Run(run func(ctx context.Context, accountID string)) *CreditService_GetBalance_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CreditService_GetBalance_Call) Return(_a0 int64, _a1 error) *CreditService_GetBalance_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CreditService_GetBalance_Call) RunAndReturn(run func(context.Context, string) (int64, error)) *CreditService_GetBalance_Call {
	_c.Call.Return(run)
	return _c
}

// GetTotalDebitedAmount provides a mock function with given fields: ctx, accountID
func (_m *CreditService) GetTotalDebitedAmount(ctx context.Context, accountID string) (int64, error) {
	ret := _m.Called(ctx, accountID)

	if len(ret) == 0 {
		panic("no return value specified for GetTotalDebitedAmount")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (int64, error)); ok {
		return rf(ctx, accountID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) int64); ok {
		r0 = rf(ctx, accountID)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, accountID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreditService_GetTotalDebitedAmount_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTotalDebitedAmount'
type CreditService_GetTotalDebitedAmount_Call struct {
	*mock.Call
}

// GetTotalDebitedAmount is a helper method to define mock.On call
//   - ctx context.Context
//   - accountID string
func (_e *CreditService_Expecter) GetTotalDebitedAmount(ctx interface{}, accountID interface{}) *CreditService_GetTotalDebitedAmount_Call {
	return &CreditService_GetTotalDebitedAmount_Call{Call: _e.mock.On("GetTotalDebitedAmount", ctx, accountID)}
}

func (_c *CreditService_GetTotalDebitedAmount_Call) Run(run func(ctx context.Context, accountID string)) *CreditService_GetTotalDebitedAmount_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CreditService_GetTotalDebitedAmount_Call) Return(_a0 int64, _a1 error) *CreditService_GetTotalDebitedAmount_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CreditService_GetTotalDebitedAmount_Call) RunAndReturn(run func(context.Context, string) (int64, error)) *CreditService_GetTotalDebitedAmount_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: ctx, filter
func (_m *CreditService) List(ctx context.Context, filter credit.Filter) ([]credit.Transaction, error) {
	ret := _m.Called(ctx, filter)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []credit.Transaction
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, credit.Filter) ([]credit.Transaction, error)); ok {
		return rf(ctx, filter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, credit.Filter) []credit.Transaction); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]credit.Transaction)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, credit.Filter) error); ok {
		r1 = rf(ctx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreditService_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type CreditService_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
//   - filter credit.Filter
func (_e *CreditService_Expecter) List(ctx interface{}, filter interface{}) *CreditService_List_Call {
	return &CreditService_List_Call{Call: _e.mock.On("List", ctx, filter)}
}

func (_c *CreditService_List_Call) Run(run func(ctx context.Context, filter credit.Filter)) *CreditService_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(credit.Filter))
	})
	return _c
}

func (_c *CreditService_List_Call) Return(_a0 []credit.Transaction, _a1 error) *CreditService_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CreditService_List_Call) RunAndReturn(run func(context.Context, credit.Filter) ([]credit.Transaction, error)) *CreditService_List_Call {
	_c.Call.Return(run)
	return _c
}

// NewCreditService creates a new instance of CreditService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCreditService(t interface {
	mock.TestingT
	Cleanup(func())
}) *CreditService {
	mock := &CreditService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
