// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	product "github.com/raystack/frontier/billing/product"
	mock "github.com/stretchr/testify/mock"
)

// ProductService is an autogenerated mock type for the ProductService type
type ProductService struct {
	mock.Mock
}

type ProductService_Expecter struct {
	mock *mock.Mock
}

func (_m *ProductService) EXPECT() *ProductService_Expecter {
	return &ProductService_Expecter{mock: &_m.Mock}
}

// GetByProviderID provides a mock function with given fields: ctx, id
func (_m *ProductService) GetByProviderID(ctx context.Context, id string) (product.Product, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetByProviderID")
	}

	var r0 product.Product
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (product.Product, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) product.Product); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(product.Product)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProductService_GetByProviderID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByProviderID'
type ProductService_GetByProviderID_Call struct {
	*mock.Call
}

// GetByProviderID is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *ProductService_Expecter) GetByProviderID(ctx interface{}, id interface{}) *ProductService_GetByProviderID_Call {
	return &ProductService_GetByProviderID_Call{Call: _e.mock.On("GetByProviderID", ctx, id)}
}

func (_c *ProductService_GetByProviderID_Call) Run(run func(ctx context.Context, id string)) *ProductService_GetByProviderID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *ProductService_GetByProviderID_Call) Return(_a0 product.Product, _a1 error) *ProductService_GetByProviderID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProductService_GetByProviderID_Call) RunAndReturn(run func(context.Context, string) (product.Product, error)) *ProductService_GetByProviderID_Call {
	_c.Call.Return(run)
	return _c
}

// NewProductService creates a new instance of ProductService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProductService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ProductService {
	mock := &ProductService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
