// Code generated by mockery v2.40.2. DO NOT EDIT.

package mocks

import (
	context "context"

	kyc "github.com/raystack/frontier/core/kyc"
	mock "github.com/stretchr/testify/mock"
)

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

type Repository_Expecter struct {
	mock *mock.Mock
}

func (_m *Repository) EXPECT() *Repository_Expecter {
	return &Repository_Expecter{mock: &_m.Mock}
}

// GetByOrgID provides a mock function with given fields: _a0, _a1
func (_m *Repository) GetByOrgID(_a0 context.Context, _a1 string) (kyc.KYC, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for GetByOrgID")
	}

	var r0 kyc.KYC
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (kyc.KYC, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) kyc.KYC); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(kyc.KYC)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Repository_GetByOrgID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByOrgID'
type Repository_GetByOrgID_Call struct {
	*mock.Call
}

// GetByOrgID is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
func (_e *Repository_Expecter) GetByOrgID(_a0 interface{}, _a1 interface{}) *Repository_GetByOrgID_Call {
	return &Repository_GetByOrgID_Call{Call: _e.mock.On("GetByOrgID", _a0, _a1)}
}

func (_c *Repository_GetByOrgID_Call) Run(run func(_a0 context.Context, _a1 string)) *Repository_GetByOrgID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Repository_GetByOrgID_Call) Return(_a0 kyc.KYC, _a1 error) *Repository_GetByOrgID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Repository_GetByOrgID_Call) RunAndReturn(run func(context.Context, string) (kyc.KYC, error)) *Repository_GetByOrgID_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: _a0
func (_m *Repository) List(_a0 context.Context) ([]kyc.KYC, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []kyc.KYC
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]kyc.KYC, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []kyc.KYC); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]kyc.KYC)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Repository_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type Repository_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *Repository_Expecter) List(_a0 interface{}) *Repository_List_Call {
	return &Repository_List_Call{Call: _e.mock.On("List", _a0)}
}

func (_c *Repository_List_Call) Run(run func(_a0 context.Context)) *Repository_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Repository_List_Call) Return(_a0 []kyc.KYC, _a1 error) *Repository_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Repository_List_Call) RunAndReturn(run func(context.Context) ([]kyc.KYC, error)) *Repository_List_Call {
	_c.Call.Return(run)
	return _c
}

// Upsert provides a mock function with given fields: _a0, _a1
func (_m *Repository) Upsert(_a0 context.Context, _a1 kyc.KYC) (kyc.KYC, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Upsert")
	}

	var r0 kyc.KYC
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, kyc.KYC) (kyc.KYC, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, kyc.KYC) kyc.KYC); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(kyc.KYC)
	}

	if rf, ok := ret.Get(1).(func(context.Context, kyc.KYC) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Repository_Upsert_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Upsert'
type Repository_Upsert_Call struct {
	*mock.Call
}

// Upsert is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 kyc.KYC
func (_e *Repository_Expecter) Upsert(_a0 interface{}, _a1 interface{}) *Repository_Upsert_Call {
	return &Repository_Upsert_Call{Call: _e.mock.On("Upsert", _a0, _a1)}
}

func (_c *Repository_Upsert_Call) Run(run func(_a0 context.Context, _a1 kyc.KYC)) *Repository_Upsert_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(kyc.KYC))
	})
	return _c
}

func (_c *Repository_Upsert_Call) Return(_a0 kyc.KYC, _a1 error) *Repository_Upsert_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Repository_Upsert_Call) RunAndReturn(run func(context.Context, kyc.KYC) (kyc.KYC, error)) *Repository_Upsert_Call {
	_c.Call.Return(run)
	return _c
}

// NewRepository creates a new instance of Repository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *Repository {
	mock := &Repository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
