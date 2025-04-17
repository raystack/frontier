// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	prospect "github.com/raystack/frontier/core/prospect"
	mock "github.com/stretchr/testify/mock"

	rql "github.com/raystack/salt/rql"
)

// ProspectService is an autogenerated mock type for the ProspectService type
type ProspectService struct {
	mock.Mock
}

type ProspectService_Expecter struct {
	mock *mock.Mock
}

func (_m *ProspectService) EXPECT() *ProspectService_Expecter {
	return &ProspectService_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, _a1
func (_m *ProspectService) Create(ctx context.Context, _a1 prospect.Prospect) (prospect.Prospect, error) {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 prospect.Prospect
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, prospect.Prospect) (prospect.Prospect, error)); ok {
		return rf(ctx, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, prospect.Prospect) prospect.Prospect); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Get(0).(prospect.Prospect)
	}

	if rf, ok := ret.Get(1).(func(context.Context, prospect.Prospect) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProspectService_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type ProspectService_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 prospect.Prospect
func (_e *ProspectService_Expecter) Create(ctx interface{}, _a1 interface{}) *ProspectService_Create_Call {
	return &ProspectService_Create_Call{Call: _e.mock.On("Create", ctx, _a1)}
}

func (_c *ProspectService_Create_Call) Run(run func(ctx context.Context, _a1 prospect.Prospect)) *ProspectService_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(prospect.Prospect))
	})
	return _c
}

func (_c *ProspectService_Create_Call) Return(_a0 prospect.Prospect, _a1 error) *ProspectService_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProspectService_Create_Call) RunAndReturn(run func(context.Context, prospect.Prospect) (prospect.Prospect, error)) *ProspectService_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: ctx, prospectId
func (_m *ProspectService) Delete(ctx context.Context, prospectId string) error {
	ret := _m.Called(ctx, prospectId)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, prospectId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ProspectService_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type ProspectService_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - prospectId string
func (_e *ProspectService_Expecter) Delete(ctx interface{}, prospectId interface{}) *ProspectService_Delete_Call {
	return &ProspectService_Delete_Call{Call: _e.mock.On("Delete", ctx, prospectId)}
}

func (_c *ProspectService_Delete_Call) Run(run func(ctx context.Context, prospectId string)) *ProspectService_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *ProspectService_Delete_Call) Return(_a0 error) *ProspectService_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ProspectService_Delete_Call) RunAndReturn(run func(context.Context, string) error) *ProspectService_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, prospectId
func (_m *ProspectService) Get(ctx context.Context, prospectId string) (prospect.Prospect, error) {
	ret := _m.Called(ctx, prospectId)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 prospect.Prospect
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (prospect.Prospect, error)); ok {
		return rf(ctx, prospectId)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) prospect.Prospect); ok {
		r0 = rf(ctx, prospectId)
	} else {
		r0 = ret.Get(0).(prospect.Prospect)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, prospectId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProspectService_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type ProspectService_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - prospectId string
func (_e *ProspectService_Expecter) Get(ctx interface{}, prospectId interface{}) *ProspectService_Get_Call {
	return &ProspectService_Get_Call{Call: _e.mock.On("Get", ctx, prospectId)}
}

func (_c *ProspectService_Get_Call) Run(run func(ctx context.Context, prospectId string)) *ProspectService_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *ProspectService_Get_Call) Return(_a0 prospect.Prospect, _a1 error) *ProspectService_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProspectService_Get_Call) RunAndReturn(run func(context.Context, string) (prospect.Prospect, error)) *ProspectService_Get_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: ctx, query
func (_m *ProspectService) List(ctx context.Context, query *rql.Query) (prospect.ListProspects, error) {
	ret := _m.Called(ctx, query)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 prospect.ListProspects
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *rql.Query) (prospect.ListProspects, error)); ok {
		return rf(ctx, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *rql.Query) prospect.ListProspects); ok {
		r0 = rf(ctx, query)
	} else {
		r0 = ret.Get(0).(prospect.ListProspects)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *rql.Query) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProspectService_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type ProspectService_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
//   - query *rql.Query
func (_e *ProspectService_Expecter) List(ctx interface{}, query interface{}) *ProspectService_List_Call {
	return &ProspectService_List_Call{Call: _e.mock.On("List", ctx, query)}
}

func (_c *ProspectService_List_Call) Run(run func(ctx context.Context, query *rql.Query)) *ProspectService_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*rql.Query))
	})
	return _c
}

func (_c *ProspectService_List_Call) Return(_a0 prospect.ListProspects, _a1 error) *ProspectService_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProspectService_List_Call) RunAndReturn(run func(context.Context, *rql.Query) (prospect.ListProspects, error)) *ProspectService_List_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: ctx, _a1
func (_m *ProspectService) Update(ctx context.Context, _a1 prospect.Prospect) (prospect.Prospect, error) {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 prospect.Prospect
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, prospect.Prospect) (prospect.Prospect, error)); ok {
		return rf(ctx, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, prospect.Prospect) prospect.Prospect); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Get(0).(prospect.Prospect)
	}

	if rf, ok := ret.Get(1).(func(context.Context, prospect.Prospect) error); ok {
		r1 = rf(ctx, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProspectService_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type ProspectService_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 prospect.Prospect
func (_e *ProspectService_Expecter) Update(ctx interface{}, _a1 interface{}) *ProspectService_Update_Call {
	return &ProspectService_Update_Call{Call: _e.mock.On("Update", ctx, _a1)}
}

func (_c *ProspectService_Update_Call) Run(run func(ctx context.Context, _a1 prospect.Prospect)) *ProspectService_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(prospect.Prospect))
	})
	return _c
}

func (_c *ProspectService_Update_Call) Return(_a0 prospect.Prospect, _a1 error) *ProspectService_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProspectService_Update_Call) RunAndReturn(run func(context.Context, prospect.Prospect) (prospect.Prospect, error)) *ProspectService_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewProspectService creates a new instance of ProspectService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProspectService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ProspectService {
	mock := &ProspectService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
