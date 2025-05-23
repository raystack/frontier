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

// Delete provides a mock function with given fields: ctx, id
func (_m *CustomerService) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CustomerService_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type CustomerService_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *CustomerService_Expecter) Delete(ctx interface{}, id interface{}) *CustomerService_Delete_Call {
	return &CustomerService_Delete_Call{Call: _e.mock.On("Delete", ctx, id)}
}

func (_c *CustomerService_Delete_Call) Run(run func(ctx context.Context, id string)) *CustomerService_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CustomerService_Delete_Call) Return(_a0 error) *CustomerService_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CustomerService_Delete_Call) RunAndReturn(run func(context.Context, string) error) *CustomerService_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Disable provides a mock function with given fields: ctx, id
func (_m *CustomerService) Disable(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Disable")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CustomerService_Disable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Disable'
type CustomerService_Disable_Call struct {
	*mock.Call
}

// Disable is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *CustomerService_Expecter) Disable(ctx interface{}, id interface{}) *CustomerService_Disable_Call {
	return &CustomerService_Disable_Call{Call: _e.mock.On("Disable", ctx, id)}
}

func (_c *CustomerService_Disable_Call) Run(run func(ctx context.Context, id string)) *CustomerService_Disable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CustomerService_Disable_Call) Return(_a0 error) *CustomerService_Disable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CustomerService_Disable_Call) RunAndReturn(run func(context.Context, string) error) *CustomerService_Disable_Call {
	_c.Call.Return(run)
	return _c
}

// Enable provides a mock function with given fields: ctx, id
func (_m *CustomerService) Enable(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Enable")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CustomerService_Enable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Enable'
type CustomerService_Enable_Call struct {
	*mock.Call
}

// Enable is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *CustomerService_Expecter) Enable(ctx interface{}, id interface{}) *CustomerService_Enable_Call {
	return &CustomerService_Enable_Call{Call: _e.mock.On("Enable", ctx, id)}
}

func (_c *CustomerService_Enable_Call) Run(run func(ctx context.Context, id string)) *CustomerService_Enable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CustomerService_Enable_Call) Return(_a0 error) *CustomerService_Enable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CustomerService_Enable_Call) RunAndReturn(run func(context.Context, string) error) *CustomerService_Enable_Call {
	_c.Call.Return(run)
	return _c
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *CustomerService) GetByID(ctx context.Context, id string) (customer.Customer, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetByID")
	}

	var r0 customer.Customer
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (customer.Customer, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) customer.Customer); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(customer.Customer)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CustomerService_GetByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByID'
type CustomerService_GetByID_Call struct {
	*mock.Call
}

// GetByID is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *CustomerService_Expecter) GetByID(ctx interface{}, id interface{}) *CustomerService_GetByID_Call {
	return &CustomerService_GetByID_Call{Call: _e.mock.On("GetByID", ctx, id)}
}

func (_c *CustomerService_GetByID_Call) Run(run func(ctx context.Context, id string)) *CustomerService_GetByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CustomerService_GetByID_Call) Return(_a0 customer.Customer, _a1 error) *CustomerService_GetByID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CustomerService_GetByID_Call) RunAndReturn(run func(context.Context, string) (customer.Customer, error)) *CustomerService_GetByID_Call {
	_c.Call.Return(run)
	return _c
}

// GetDetails provides a mock function with given fields: ctx, customerID
func (_m *CustomerService) GetDetails(ctx context.Context, customerID string) (customer.Details, error) {
	ret := _m.Called(ctx, customerID)

	if len(ret) == 0 {
		panic("no return value specified for GetDetails")
	}

	var r0 customer.Details
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (customer.Details, error)); ok {
		return rf(ctx, customerID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) customer.Details); ok {
		r0 = rf(ctx, customerID)
	} else {
		r0 = ret.Get(0).(customer.Details)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, customerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CustomerService_GetDetails_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDetails'
type CustomerService_GetDetails_Call struct {
	*mock.Call
}

// GetDetails is a helper method to define mock.On call
//   - ctx context.Context
//   - customerID string
func (_e *CustomerService_Expecter) GetDetails(ctx interface{}, customerID interface{}) *CustomerService_GetDetails_Call {
	return &CustomerService_GetDetails_Call{Call: _e.mock.On("GetDetails", ctx, customerID)}
}

func (_c *CustomerService_GetDetails_Call) Run(run func(ctx context.Context, customerID string)) *CustomerService_GetDetails_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CustomerService_GetDetails_Call) Return(_a0 customer.Details, _a1 error) *CustomerService_GetDetails_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CustomerService_GetDetails_Call) RunAndReturn(run func(context.Context, string) (customer.Details, error)) *CustomerService_GetDetails_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: ctx, filter
func (_m *CustomerService) List(ctx context.Context, filter customer.Filter) ([]customer.Customer, error) {
	ret := _m.Called(ctx, filter)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []customer.Customer
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, customer.Filter) ([]customer.Customer, error)); ok {
		return rf(ctx, filter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, customer.Filter) []customer.Customer); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]customer.Customer)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, customer.Filter) error); ok {
		r1 = rf(ctx, filter)
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
//   - filter customer.Filter
func (_e *CustomerService_Expecter) List(ctx interface{}, filter interface{}) *CustomerService_List_Call {
	return &CustomerService_List_Call{Call: _e.mock.On("List", ctx, filter)}
}

func (_c *CustomerService_List_Call) Run(run func(ctx context.Context, filter customer.Filter)) *CustomerService_List_Call {
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

// ListPaymentMethods provides a mock function with given fields: ctx, id
func (_m *CustomerService) ListPaymentMethods(ctx context.Context, id string) ([]customer.PaymentMethod, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for ListPaymentMethods")
	}

	var r0 []customer.PaymentMethod
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]customer.PaymentMethod, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []customer.PaymentMethod); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]customer.PaymentMethod)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CustomerService_ListPaymentMethods_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListPaymentMethods'
type CustomerService_ListPaymentMethods_Call struct {
	*mock.Call
}

// ListPaymentMethods is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *CustomerService_Expecter) ListPaymentMethods(ctx interface{}, id interface{}) *CustomerService_ListPaymentMethods_Call {
	return &CustomerService_ListPaymentMethods_Call{Call: _e.mock.On("ListPaymentMethods", ctx, id)}
}

func (_c *CustomerService_ListPaymentMethods_Call) Run(run func(ctx context.Context, id string)) *CustomerService_ListPaymentMethods_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CustomerService_ListPaymentMethods_Call) Return(_a0 []customer.PaymentMethod, _a1 error) *CustomerService_ListPaymentMethods_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CustomerService_ListPaymentMethods_Call) RunAndReturn(run func(context.Context, string) ([]customer.PaymentMethod, error)) *CustomerService_ListPaymentMethods_Call {
	_c.Call.Return(run)
	return _c
}

// RegisterToProviderIfRequired provides a mock function with given fields: ctx, customerID
func (_m *CustomerService) RegisterToProviderIfRequired(ctx context.Context, customerID string) (customer.Customer, error) {
	ret := _m.Called(ctx, customerID)

	if len(ret) == 0 {
		panic("no return value specified for RegisterToProviderIfRequired")
	}

	var r0 customer.Customer
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (customer.Customer, error)); ok {
		return rf(ctx, customerID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) customer.Customer); ok {
		r0 = rf(ctx, customerID)
	} else {
		r0 = ret.Get(0).(customer.Customer)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, customerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CustomerService_RegisterToProviderIfRequired_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RegisterToProviderIfRequired'
type CustomerService_RegisterToProviderIfRequired_Call struct {
	*mock.Call
}

// RegisterToProviderIfRequired is a helper method to define mock.On call
//   - ctx context.Context
//   - customerID string
func (_e *CustomerService_Expecter) RegisterToProviderIfRequired(ctx interface{}, customerID interface{}) *CustomerService_RegisterToProviderIfRequired_Call {
	return &CustomerService_RegisterToProviderIfRequired_Call{Call: _e.mock.On("RegisterToProviderIfRequired", ctx, customerID)}
}

func (_c *CustomerService_RegisterToProviderIfRequired_Call) Run(run func(ctx context.Context, customerID string)) *CustomerService_RegisterToProviderIfRequired_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CustomerService_RegisterToProviderIfRequired_Call) Return(_a0 customer.Customer, _a1 error) *CustomerService_RegisterToProviderIfRequired_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CustomerService_RegisterToProviderIfRequired_Call) RunAndReturn(run func(context.Context, string) (customer.Customer, error)) *CustomerService_RegisterToProviderIfRequired_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: ctx, _a1
func (_m *CustomerService) Update(ctx context.Context, _a1 customer.Customer) (customer.Customer, error) {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 customer.Customer
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, customer.Customer) (customer.Customer, error)); ok {
		return rf(ctx, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, customer.Customer) customer.Customer); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Get(0).(customer.Customer)
	}

	if rf, ok := ret.Get(1).(func(context.Context, customer.Customer) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CustomerService_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type CustomerService_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 customer.Customer
func (_e *CustomerService_Expecter) Update(ctx interface{}, _a1 interface{}) *CustomerService_Update_Call {
	return &CustomerService_Update_Call{Call: _e.mock.On("Update", ctx, _a1)}
}

func (_c *CustomerService_Update_Call) Run(run func(ctx context.Context, _a1 customer.Customer)) *CustomerService_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(customer.Customer))
	})
	return _c
}

func (_c *CustomerService_Update_Call) Return(_a0 customer.Customer, _a1 error) *CustomerService_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CustomerService_Update_Call) RunAndReturn(run func(context.Context, customer.Customer) (customer.Customer, error)) *CustomerService_Update_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateCreditMinByID provides a mock function with given fields: ctx, customerID, limit
func (_m *CustomerService) UpdateCreditMinByID(ctx context.Context, customerID string, limit int64) (customer.Details, error) {
	ret := _m.Called(ctx, customerID, limit)

	if len(ret) == 0 {
		panic("no return value specified for UpdateCreditMinByID")
	}

	var r0 customer.Details
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int64) (customer.Details, error)); ok {
		return rf(ctx, customerID, limit)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, int64) customer.Details); ok {
		r0 = rf(ctx, customerID, limit)
	} else {
		r0 = ret.Get(0).(customer.Details)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, int64) error); ok {
		r1 = rf(ctx, customerID, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CustomerService_UpdateCreditMinByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateCreditMinByID'
type CustomerService_UpdateCreditMinByID_Call struct {
	*mock.Call
}

// UpdateCreditMinByID is a helper method to define mock.On call
//   - ctx context.Context
//   - customerID string
//   - limit int64
func (_e *CustomerService_Expecter) UpdateCreditMinByID(ctx interface{}, customerID interface{}, limit interface{}) *CustomerService_UpdateCreditMinByID_Call {
	return &CustomerService_UpdateCreditMinByID_Call{Call: _e.mock.On("UpdateCreditMinByID", ctx, customerID, limit)}
}

func (_c *CustomerService_UpdateCreditMinByID_Call) Run(run func(ctx context.Context, customerID string, limit int64)) *CustomerService_UpdateCreditMinByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(int64))
	})
	return _c
}

func (_c *CustomerService_UpdateCreditMinByID_Call) Return(_a0 customer.Details, _a1 error) *CustomerService_UpdateCreditMinByID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CustomerService_UpdateCreditMinByID_Call) RunAndReturn(run func(context.Context, string, int64) (customer.Details, error)) *CustomerService_UpdateCreditMinByID_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateDetails provides a mock function with given fields: ctx, customerID, details
func (_m *CustomerService) UpdateDetails(ctx context.Context, customerID string, details customer.Details) (customer.Details, error) {
	ret := _m.Called(ctx, customerID, details)

	if len(ret) == 0 {
		panic("no return value specified for UpdateDetails")
	}

	var r0 customer.Details
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, customer.Details) (customer.Details, error)); ok {
		return rf(ctx, customerID, details)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, customer.Details) customer.Details); ok {
		r0 = rf(ctx, customerID, details)
	} else {
		r0 = ret.Get(0).(customer.Details)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, customer.Details) error); ok {
		r1 = rf(ctx, customerID, details)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CustomerService_UpdateDetails_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateDetails'
type CustomerService_UpdateDetails_Call struct {
	*mock.Call
}

// UpdateDetails is a helper method to define mock.On call
//   - ctx context.Context
//   - customerID string
//   - details customer.Details
func (_e *CustomerService_Expecter) UpdateDetails(ctx interface{}, customerID interface{}, details interface{}) *CustomerService_UpdateDetails_Call {
	return &CustomerService_UpdateDetails_Call{Call: _e.mock.On("UpdateDetails", ctx, customerID, details)}
}

func (_c *CustomerService_UpdateDetails_Call) Run(run func(ctx context.Context, customerID string, details customer.Details)) *CustomerService_UpdateDetails_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(customer.Details))
	})
	return _c
}

func (_c *CustomerService_UpdateDetails_Call) Return(_a0 customer.Details, _a1 error) *CustomerService_UpdateDetails_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CustomerService_UpdateDetails_Call) RunAndReturn(run func(context.Context, string, customer.Details) (customer.Details, error)) *CustomerService_UpdateDetails_Call {
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
