// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	authenticate "github.com/raystack/frontier/core/authenticate"

	mock "github.com/stretchr/testify/mock"

	organization "github.com/raystack/frontier/core/organization"
)

// OrganizationService is an autogenerated mock type for the OrganizationService type
type OrganizationService struct {
	mock.Mock
}

type OrganizationService_Expecter struct {
	mock *mock.Mock
}

func (_m *OrganizationService) EXPECT() *OrganizationService_Expecter {
	return &OrganizationService_Expecter{mock: &_m.Mock}
}

// AddUsers provides a mock function with given fields: ctx, orgID, userID
func (_m *OrganizationService) AddUsers(ctx context.Context, orgID string, userID []string) error {
	ret := _m.Called(ctx, orgID, userID)

	if len(ret) == 0 {
		panic("no return value specified for AddUsers")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) error); ok {
		r0 = rf(ctx, orgID, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OrganizationService_AddUsers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddUsers'
type OrganizationService_AddUsers_Call struct {
	*mock.Call
}

// AddUsers is a helper method to define mock.On call
//   - ctx context.Context
//   - orgID string
//   - userID []string
func (_e *OrganizationService_Expecter) AddUsers(ctx interface{}, orgID interface{}, userID interface{}) *OrganizationService_AddUsers_Call {
	return &OrganizationService_AddUsers_Call{Call: _e.mock.On("AddUsers", ctx, orgID, userID)}
}

func (_c *OrganizationService_AddUsers_Call) Run(run func(ctx context.Context, orgID string, userID []string)) *OrganizationService_AddUsers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].([]string))
	})
	return _c
}

func (_c *OrganizationService_AddUsers_Call) Return(_a0 error) *OrganizationService_AddUsers_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *OrganizationService_AddUsers_Call) RunAndReturn(run func(context.Context, string, []string) error) *OrganizationService_AddUsers_Call {
	_c.Call.Return(run)
	return _c
}

// AdminCreate provides a mock function with given fields: ctx, org, ownerEmail
func (_m *OrganizationService) AdminCreate(ctx context.Context, org organization.Organization, ownerEmail string) (organization.Organization, error) {
	ret := _m.Called(ctx, org, ownerEmail)

	if len(ret) == 0 {
		panic("no return value specified for AdminCreate")
	}

	var r0 organization.Organization
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, organization.Organization, string) (organization.Organization, error)); ok {
		return rf(ctx, org, ownerEmail)
	}
	if rf, ok := ret.Get(0).(func(context.Context, organization.Organization, string) organization.Organization); ok {
		r0 = rf(ctx, org, ownerEmail)
	} else {
		r0 = ret.Get(0).(organization.Organization)
	}

	if rf, ok := ret.Get(1).(func(context.Context, organization.Organization, string) error); ok {
		r1 = rf(ctx, org, ownerEmail)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OrganizationService_AdminCreate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AdminCreate'
type OrganizationService_AdminCreate_Call struct {
	*mock.Call
}

// AdminCreate is a helper method to define mock.On call
//   - ctx context.Context
//   - org organization.Organization
//   - ownerEmail string
func (_e *OrganizationService_Expecter) AdminCreate(ctx interface{}, org interface{}, ownerEmail interface{}) *OrganizationService_AdminCreate_Call {
	return &OrganizationService_AdminCreate_Call{Call: _e.mock.On("AdminCreate", ctx, org, ownerEmail)}
}

func (_c *OrganizationService_AdminCreate_Call) Run(run func(ctx context.Context, org organization.Organization, ownerEmail string)) *OrganizationService_AdminCreate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(organization.Organization), args[2].(string))
	})
	return _c
}

func (_c *OrganizationService_AdminCreate_Call) Return(_a0 organization.Organization, _a1 error) *OrganizationService_AdminCreate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OrganizationService_AdminCreate_Call) RunAndReturn(run func(context.Context, organization.Organization, string) (organization.Organization, error)) *OrganizationService_AdminCreate_Call {
	_c.Call.Return(run)
	return _c
}

// Create provides a mock function with given fields: ctx, org
func (_m *OrganizationService) Create(ctx context.Context, org organization.Organization) (organization.Organization, error) {
	ret := _m.Called(ctx, org)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 organization.Organization
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, organization.Organization) (organization.Organization, error)); ok {
		return rf(ctx, org)
	}
	if rf, ok := ret.Get(0).(func(context.Context, organization.Organization) organization.Organization); ok {
		r0 = rf(ctx, org)
	} else {
		r0 = ret.Get(0).(organization.Organization)
	}

	if rf, ok := ret.Get(1).(func(context.Context, organization.Organization) error); ok {
		r1 = rf(ctx, org)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OrganizationService_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type OrganizationService_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - org organization.Organization
func (_e *OrganizationService_Expecter) Create(ctx interface{}, org interface{}) *OrganizationService_Create_Call {
	return &OrganizationService_Create_Call{Call: _e.mock.On("Create", ctx, org)}
}

func (_c *OrganizationService_Create_Call) Run(run func(ctx context.Context, org organization.Organization)) *OrganizationService_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(organization.Organization))
	})
	return _c
}

func (_c *OrganizationService_Create_Call) Return(_a0 organization.Organization, _a1 error) *OrganizationService_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OrganizationService_Create_Call) RunAndReturn(run func(context.Context, organization.Organization) (organization.Organization, error)) *OrganizationService_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Disable provides a mock function with given fields: ctx, id
func (_m *OrganizationService) Disable(ctx context.Context, id string) error {
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

// OrganizationService_Disable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Disable'
type OrganizationService_Disable_Call struct {
	*mock.Call
}

// Disable is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *OrganizationService_Expecter) Disable(ctx interface{}, id interface{}) *OrganizationService_Disable_Call {
	return &OrganizationService_Disable_Call{Call: _e.mock.On("Disable", ctx, id)}
}

func (_c *OrganizationService_Disable_Call) Run(run func(ctx context.Context, id string)) *OrganizationService_Disable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *OrganizationService_Disable_Call) Return(_a0 error) *OrganizationService_Disable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *OrganizationService_Disable_Call) RunAndReturn(run func(context.Context, string) error) *OrganizationService_Disable_Call {
	_c.Call.Return(run)
	return _c
}

// Enable provides a mock function with given fields: ctx, id
func (_m *OrganizationService) Enable(ctx context.Context, id string) error {
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

// OrganizationService_Enable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Enable'
type OrganizationService_Enable_Call struct {
	*mock.Call
}

// Enable is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *OrganizationService_Expecter) Enable(ctx interface{}, id interface{}) *OrganizationService_Enable_Call {
	return &OrganizationService_Enable_Call{Call: _e.mock.On("Enable", ctx, id)}
}

func (_c *OrganizationService_Enable_Call) Run(run func(ctx context.Context, id string)) *OrganizationService_Enable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *OrganizationService_Enable_Call) Return(_a0 error) *OrganizationService_Enable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *OrganizationService_Enable_Call) RunAndReturn(run func(context.Context, string) error) *OrganizationService_Enable_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, idOrSlug
func (_m *OrganizationService) Get(ctx context.Context, idOrSlug string) (organization.Organization, error) {
	ret := _m.Called(ctx, idOrSlug)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 organization.Organization
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (organization.Organization, error)); ok {
		return rf(ctx, idOrSlug)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) organization.Organization); ok {
		r0 = rf(ctx, idOrSlug)
	} else {
		r0 = ret.Get(0).(organization.Organization)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, idOrSlug)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OrganizationService_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type OrganizationService_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - idOrSlug string
func (_e *OrganizationService_Expecter) Get(ctx interface{}, idOrSlug interface{}) *OrganizationService_Get_Call {
	return &OrganizationService_Get_Call{Call: _e.mock.On("Get", ctx, idOrSlug)}
}

func (_c *OrganizationService_Get_Call) Run(run func(ctx context.Context, idOrSlug string)) *OrganizationService_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *OrganizationService_Get_Call) Return(_a0 organization.Organization, _a1 error) *OrganizationService_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OrganizationService_Get_Call) RunAndReturn(run func(context.Context, string) (organization.Organization, error)) *OrganizationService_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetRaw provides a mock function with given fields: ctx, idOrSlug
func (_m *OrganizationService) GetRaw(ctx context.Context, idOrSlug string) (organization.Organization, error) {
	ret := _m.Called(ctx, idOrSlug)

	if len(ret) == 0 {
		panic("no return value specified for GetRaw")
	}

	var r0 organization.Organization
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (organization.Organization, error)); ok {
		return rf(ctx, idOrSlug)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) organization.Organization); ok {
		r0 = rf(ctx, idOrSlug)
	} else {
		r0 = ret.Get(0).(organization.Organization)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, idOrSlug)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OrganizationService_GetRaw_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRaw'
type OrganizationService_GetRaw_Call struct {
	*mock.Call
}

// GetRaw is a helper method to define mock.On call
//   - ctx context.Context
//   - idOrSlug string
func (_e *OrganizationService_Expecter) GetRaw(ctx interface{}, idOrSlug interface{}) *OrganizationService_GetRaw_Call {
	return &OrganizationService_GetRaw_Call{Call: _e.mock.On("GetRaw", ctx, idOrSlug)}
}

func (_c *OrganizationService_GetRaw_Call) Run(run func(ctx context.Context, idOrSlug string)) *OrganizationService_GetRaw_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *OrganizationService_GetRaw_Call) Return(_a0 organization.Organization, _a1 error) *OrganizationService_GetRaw_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OrganizationService_GetRaw_Call) RunAndReturn(run func(context.Context, string) (organization.Organization, error)) *OrganizationService_GetRaw_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: ctx, f
func (_m *OrganizationService) List(ctx context.Context, f organization.Filter) ([]organization.Organization, error) {
	ret := _m.Called(ctx, f)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []organization.Organization
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, organization.Filter) ([]organization.Organization, error)); ok {
		return rf(ctx, f)
	}
	if rf, ok := ret.Get(0).(func(context.Context, organization.Filter) []organization.Organization); ok {
		r0 = rf(ctx, f)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]organization.Organization)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, organization.Filter) error); ok {
		r1 = rf(ctx, f)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OrganizationService_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type OrganizationService_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
//   - f organization.Filter
func (_e *OrganizationService_Expecter) List(ctx interface{}, f interface{}) *OrganizationService_List_Call {
	return &OrganizationService_List_Call{Call: _e.mock.On("List", ctx, f)}
}

func (_c *OrganizationService_List_Call) Run(run func(ctx context.Context, f organization.Filter)) *OrganizationService_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(organization.Filter))
	})
	return _c
}

func (_c *OrganizationService_List_Call) Return(_a0 []organization.Organization, _a1 error) *OrganizationService_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OrganizationService_List_Call) RunAndReturn(run func(context.Context, organization.Filter) ([]organization.Organization, error)) *OrganizationService_List_Call {
	_c.Call.Return(run)
	return _c
}

// ListByUser provides a mock function with given fields: ctx, principal, flt
func (_m *OrganizationService) ListByUser(ctx context.Context, principal authenticate.Principal, flt organization.Filter) ([]organization.Organization, error) {
	ret := _m.Called(ctx, principal, flt)

	if len(ret) == 0 {
		panic("no return value specified for ListByUser")
	}

	var r0 []organization.Organization
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authenticate.Principal, organization.Filter) ([]organization.Organization, error)); ok {
		return rf(ctx, principal, flt)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authenticate.Principal, organization.Filter) []organization.Organization); ok {
		r0 = rf(ctx, principal, flt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]organization.Organization)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authenticate.Principal, organization.Filter) error); ok {
		r1 = rf(ctx, principal, flt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OrganizationService_ListByUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListByUser'
type OrganizationService_ListByUser_Call struct {
	*mock.Call
}

// ListByUser is a helper method to define mock.On call
//   - ctx context.Context
//   - principal authenticate.Principal
//   - flt organization.Filter
func (_e *OrganizationService_Expecter) ListByUser(ctx interface{}, principal interface{}, flt interface{}) *OrganizationService_ListByUser_Call {
	return &OrganizationService_ListByUser_Call{Call: _e.mock.On("ListByUser", ctx, principal, flt)}
}

func (_c *OrganizationService_ListByUser_Call) Run(run func(ctx context.Context, principal authenticate.Principal, flt organization.Filter)) *OrganizationService_ListByUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(authenticate.Principal), args[2].(organization.Filter))
	})
	return _c
}

func (_c *OrganizationService_ListByUser_Call) Return(_a0 []organization.Organization, _a1 error) *OrganizationService_ListByUser_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OrganizationService_ListByUser_Call) RunAndReturn(run func(context.Context, authenticate.Principal, organization.Filter) ([]organization.Organization, error)) *OrganizationService_ListByUser_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: ctx, toUpdate
func (_m *OrganizationService) Update(ctx context.Context, toUpdate organization.Organization) (organization.Organization, error) {
	ret := _m.Called(ctx, toUpdate)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 organization.Organization
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, organization.Organization) (organization.Organization, error)); ok {
		return rf(ctx, toUpdate)
	}
	if rf, ok := ret.Get(0).(func(context.Context, organization.Organization) organization.Organization); ok {
		r0 = rf(ctx, toUpdate)
	} else {
		r0 = ret.Get(0).(organization.Organization)
	}

	if rf, ok := ret.Get(1).(func(context.Context, organization.Organization) error); ok {
		r1 = rf(ctx, toUpdate)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OrganizationService_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type OrganizationService_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - toUpdate organization.Organization
func (_e *OrganizationService_Expecter) Update(ctx interface{}, toUpdate interface{}) *OrganizationService_Update_Call {
	return &OrganizationService_Update_Call{Call: _e.mock.On("Update", ctx, toUpdate)}
}

func (_c *OrganizationService_Update_Call) Run(run func(ctx context.Context, toUpdate organization.Organization)) *OrganizationService_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(organization.Organization))
	})
	return _c
}

func (_c *OrganizationService_Update_Call) Return(_a0 organization.Organization, _a1 error) *OrganizationService_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OrganizationService_Update_Call) RunAndReturn(run func(context.Context, organization.Organization) (organization.Organization, error)) *OrganizationService_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewOrganizationService creates a new instance of OrganizationService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewOrganizationService(t interface {
	mock.TestingT
	Cleanup(func())
}) *OrganizationService {
	mock := &OrganizationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
