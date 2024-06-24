// Code generated by mockery v2.40.2. DO NOT EDIT.

package mocks

import (
	context "context"

	group "github.com/raystack/frontier/core/group"
	mock "github.com/stretchr/testify/mock"

	project "github.com/raystack/frontier/core/project"

	serviceuser "github.com/raystack/frontier/core/serviceuser"

	user "github.com/raystack/frontier/core/user"
)

// ProjectService is an autogenerated mock type for the ProjectService type
type ProjectService struct {
	mock.Mock
}

type ProjectService_Expecter struct {
	mock *mock.Mock
}

func (_m *ProjectService) EXPECT() *ProjectService_Expecter {
	return &ProjectService_Expecter{mock: &_m.Mock}
}

// AddPrincipal provides a mock function with given fields: ctx, id, roleID, principal
func (_m *ProjectService) AddPrincipal(ctx context.Context, id string, roleID string, principal project.Principal) error {
	ret := _m.Called(ctx, id, roleID, principal)

	if len(ret) == 0 {
		panic("no return value specified for AddPrincipal")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, project.Principal) error); ok {
		r0 = rf(ctx, id, roleID, principal)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ProjectService_AddPrincipal_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddPrincipal'
type ProjectService_AddPrincipal_Call struct {
	*mock.Call
}

// AddPrincipal is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - roleID string
//   - principal project.Principal
func (_e *ProjectService_Expecter) AddPrincipal(ctx interface{}, id interface{}, roleID interface{}, principal interface{}) *ProjectService_AddPrincipal_Call {
	return &ProjectService_AddPrincipal_Call{Call: _e.mock.On("AddPrincipal", ctx, id, roleID, principal)}
}

func (_c *ProjectService_AddPrincipal_Call) Run(run func(ctx context.Context, id string, roleID string, principal project.Principal)) *ProjectService_AddPrincipal_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(project.Principal))
	})
	return _c
}

func (_c *ProjectService_AddPrincipal_Call) Return(_a0 error) *ProjectService_AddPrincipal_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ProjectService_AddPrincipal_Call) RunAndReturn(run func(context.Context, string, string, project.Principal) error) *ProjectService_AddPrincipal_Call {
	_c.Call.Return(run)
	return _c
}

// Create provides a mock function with given fields: ctx, prj
func (_m *ProjectService) Create(ctx context.Context, prj project.Project) (project.Project, error) {
	ret := _m.Called(ctx, prj)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 project.Project
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, project.Project) (project.Project, error)); ok {
		return rf(ctx, prj)
	}
	if rf, ok := ret.Get(0).(func(context.Context, project.Project) project.Project); ok {
		r0 = rf(ctx, prj)
	} else {
		r0 = ret.Get(0).(project.Project)
	}

	if rf, ok := ret.Get(1).(func(context.Context, project.Project) error); ok {
		r1 = rf(ctx, prj)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProjectService_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type ProjectService_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - prj project.Project
func (_e *ProjectService_Expecter) Create(ctx interface{}, prj interface{}) *ProjectService_Create_Call {
	return &ProjectService_Create_Call{Call: _e.mock.On("Create", ctx, prj)}
}

func (_c *ProjectService_Create_Call) Run(run func(ctx context.Context, prj project.Project)) *ProjectService_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(project.Project))
	})
	return _c
}

func (_c *ProjectService_Create_Call) Return(_a0 project.Project, _a1 error) *ProjectService_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProjectService_Create_Call) RunAndReturn(run func(context.Context, project.Project) (project.Project, error)) *ProjectService_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Disable provides a mock function with given fields: ctx, id
func (_m *ProjectService) Disable(ctx context.Context, id string) error {
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

// ProjectService_Disable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Disable'
type ProjectService_Disable_Call struct {
	*mock.Call
}

// Disable is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *ProjectService_Expecter) Disable(ctx interface{}, id interface{}) *ProjectService_Disable_Call {
	return &ProjectService_Disable_Call{Call: _e.mock.On("Disable", ctx, id)}
}

func (_c *ProjectService_Disable_Call) Run(run func(ctx context.Context, id string)) *ProjectService_Disable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *ProjectService_Disable_Call) Return(_a0 error) *ProjectService_Disable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ProjectService_Disable_Call) RunAndReturn(run func(context.Context, string) error) *ProjectService_Disable_Call {
	_c.Call.Return(run)
	return _c
}

// Enable provides a mock function with given fields: ctx, id
func (_m *ProjectService) Enable(ctx context.Context, id string) error {
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

// ProjectService_Enable_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Enable'
type ProjectService_Enable_Call struct {
	*mock.Call
}

// Enable is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *ProjectService_Expecter) Enable(ctx interface{}, id interface{}) *ProjectService_Enable_Call {
	return &ProjectService_Enable_Call{Call: _e.mock.On("Enable", ctx, id)}
}

func (_c *ProjectService_Enable_Call) Run(run func(ctx context.Context, id string)) *ProjectService_Enable_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *ProjectService_Enable_Call) Return(_a0 error) *ProjectService_Enable_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ProjectService_Enable_Call) RunAndReturn(run func(context.Context, string) error) *ProjectService_Enable_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, idOrName
func (_m *ProjectService) Get(ctx context.Context, idOrName string) (project.Project, error) {
	ret := _m.Called(ctx, idOrName)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 project.Project
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (project.Project, error)); ok {
		return rf(ctx, idOrName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) project.Project); ok {
		r0 = rf(ctx, idOrName)
	} else {
		r0 = ret.Get(0).(project.Project)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, idOrName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProjectService_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type ProjectService_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - idOrName string
func (_e *ProjectService_Expecter) Get(ctx interface{}, idOrName interface{}) *ProjectService_Get_Call {
	return &ProjectService_Get_Call{Call: _e.mock.On("Get", ctx, idOrName)}
}

func (_c *ProjectService_Get_Call) Run(run func(ctx context.Context, idOrName string)) *ProjectService_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *ProjectService_Get_Call) Return(_a0 project.Project, _a1 error) *ProjectService_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProjectService_Get_Call) RunAndReturn(run func(context.Context, string) (project.Project, error)) *ProjectService_Get_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: ctx, f
func (_m *ProjectService) List(ctx context.Context, f project.Filter) ([]project.Project, error) {
	ret := _m.Called(ctx, f)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []project.Project
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, project.Filter) ([]project.Project, error)); ok {
		return rf(ctx, f)
	}
	if rf, ok := ret.Get(0).(func(context.Context, project.Filter) []project.Project); ok {
		r0 = rf(ctx, f)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]project.Project)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, project.Filter) error); ok {
		r1 = rf(ctx, f)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProjectService_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type ProjectService_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
//   - f project.Filter
func (_e *ProjectService_Expecter) List(ctx interface{}, f interface{}) *ProjectService_List_Call {
	return &ProjectService_List_Call{Call: _e.mock.On("List", ctx, f)}
}

func (_c *ProjectService_List_Call) Run(run func(ctx context.Context, f project.Filter)) *ProjectService_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(project.Filter))
	})
	return _c
}

func (_c *ProjectService_List_Call) Return(_a0 []project.Project, _a1 error) *ProjectService_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProjectService_List_Call) RunAndReturn(run func(context.Context, project.Filter) ([]project.Project, error)) *ProjectService_List_Call {
	_c.Call.Return(run)
	return _c
}

// ListByUser provides a mock function with given fields: ctx, principalID, principalType, flt
func (_m *ProjectService) ListByUser(ctx context.Context, principalID string, principalType string, flt project.Filter) ([]project.Project, error) {
	ret := _m.Called(ctx, principalID, principalType, flt)

	if len(ret) == 0 {
		panic("no return value specified for ListByUser")
	}

	var r0 []project.Project
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, project.Filter) ([]project.Project, error)); ok {
		return rf(ctx, principalID, principalType, flt)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, project.Filter) []project.Project); ok {
		r0 = rf(ctx, principalID, principalType, flt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]project.Project)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, project.Filter) error); ok {
		r1 = rf(ctx, principalID, principalType, flt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProjectService_ListByUser_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListByUser'
type ProjectService_ListByUser_Call struct {
	*mock.Call
}

// ListByUser is a helper method to define mock.On call
//   - ctx context.Context
//   - principalID string
//   - principalType string
//   - flt project.Filter
func (_e *ProjectService_Expecter) ListByUser(ctx interface{}, principalID interface{}, principalType interface{}, flt interface{}) *ProjectService_ListByUser_Call {
	return &ProjectService_ListByUser_Call{Call: _e.mock.On("ListByUser", ctx, principalID, principalType, flt)}
}

func (_c *ProjectService_ListByUser_Call) Run(run func(ctx context.Context, principalID string, principalType string, flt project.Filter)) *ProjectService_ListByUser_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(project.Filter))
	})
	return _c
}

func (_c *ProjectService_ListByUser_Call) Return(_a0 []project.Project, _a1 error) *ProjectService_ListByUser_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProjectService_ListByUser_Call) RunAndReturn(run func(context.Context, string, string, project.Filter) ([]project.Project, error)) *ProjectService_ListByUser_Call {
	_c.Call.Return(run)
	return _c
}

// ListGroups provides a mock function with given fields: ctx, id
func (_m *ProjectService) ListGroups(ctx context.Context, id string) ([]group.Group, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for ListGroups")
	}

	var r0 []group.Group
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]group.Group, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []group.Group); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]group.Group)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProjectService_ListGroups_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListGroups'
type ProjectService_ListGroups_Call struct {
	*mock.Call
}

// ListGroups is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *ProjectService_Expecter) ListGroups(ctx interface{}, id interface{}) *ProjectService_ListGroups_Call {
	return &ProjectService_ListGroups_Call{Call: _e.mock.On("ListGroups", ctx, id)}
}

func (_c *ProjectService_ListGroups_Call) Run(run func(ctx context.Context, id string)) *ProjectService_ListGroups_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *ProjectService_ListGroups_Call) Return(_a0 []group.Group, _a1 error) *ProjectService_ListGroups_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProjectService_ListGroups_Call) RunAndReturn(run func(context.Context, string) ([]group.Group, error)) *ProjectService_ListGroups_Call {
	_c.Call.Return(run)
	return _c
}

// ListServiceUsers provides a mock function with given fields: ctx, id, permissionFilter
func (_m *ProjectService) ListServiceUsers(ctx context.Context, id string, permissionFilter string) ([]serviceuser.ServiceUser, error) {
	ret := _m.Called(ctx, id, permissionFilter)

	if len(ret) == 0 {
		panic("no return value specified for ListServiceUsers")
	}

	var r0 []serviceuser.ServiceUser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) ([]serviceuser.ServiceUser, error)); ok {
		return rf(ctx, id, permissionFilter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []serviceuser.ServiceUser); ok {
		r0 = rf(ctx, id, permissionFilter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]serviceuser.ServiceUser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, permissionFilter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProjectService_ListServiceUsers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListServiceUsers'
type ProjectService_ListServiceUsers_Call struct {
	*mock.Call
}

// ListServiceUsers is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - permissionFilter string
func (_e *ProjectService_Expecter) ListServiceUsers(ctx interface{}, id interface{}, permissionFilter interface{}) *ProjectService_ListServiceUsers_Call {
	return &ProjectService_ListServiceUsers_Call{Call: _e.mock.On("ListServiceUsers", ctx, id, permissionFilter)}
}

func (_c *ProjectService_ListServiceUsers_Call) Run(run func(ctx context.Context, id string, permissionFilter string)) *ProjectService_ListServiceUsers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *ProjectService_ListServiceUsers_Call) Return(_a0 []serviceuser.ServiceUser, _a1 error) *ProjectService_ListServiceUsers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProjectService_ListServiceUsers_Call) RunAndReturn(run func(context.Context, string, string) ([]serviceuser.ServiceUser, error)) *ProjectService_ListServiceUsers_Call {
	_c.Call.Return(run)
	return _c
}

// ListUsers provides a mock function with given fields: ctx, id, permissionFilter
func (_m *ProjectService) ListUsers(ctx context.Context, id string, permissionFilter string) ([]user.User, error) {
	ret := _m.Called(ctx, id, permissionFilter)

	if len(ret) == 0 {
		panic("no return value specified for ListUsers")
	}

	var r0 []user.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) ([]user.User, error)); ok {
		return rf(ctx, id, permissionFilter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []user.User); ok {
		r0 = rf(ctx, id, permissionFilter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]user.User)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, permissionFilter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProjectService_ListUsers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListUsers'
type ProjectService_ListUsers_Call struct {
	*mock.Call
}

// ListUsers is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - permissionFilter string
func (_e *ProjectService_Expecter) ListUsers(ctx interface{}, id interface{}, permissionFilter interface{}) *ProjectService_ListUsers_Call {
	return &ProjectService_ListUsers_Call{Call: _e.mock.On("ListUsers", ctx, id, permissionFilter)}
}

func (_c *ProjectService_ListUsers_Call) Run(run func(ctx context.Context, id string, permissionFilter string)) *ProjectService_ListUsers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *ProjectService_ListUsers_Call) Return(_a0 []user.User, _a1 error) *ProjectService_ListUsers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProjectService_ListUsers_Call) RunAndReturn(run func(context.Context, string, string) ([]user.User, error)) *ProjectService_ListUsers_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: ctx, toUpdate
func (_m *ProjectService) Update(ctx context.Context, toUpdate project.Project) (project.Project, error) {
	ret := _m.Called(ctx, toUpdate)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 project.Project
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, project.Project) (project.Project, error)); ok {
		return rf(ctx, toUpdate)
	}
	if rf, ok := ret.Get(0).(func(context.Context, project.Project) project.Project); ok {
		r0 = rf(ctx, toUpdate)
	} else {
		r0 = ret.Get(0).(project.Project)
	}

	if rf, ok := ret.Get(1).(func(context.Context, project.Project) error); ok {
		r1 = rf(ctx, toUpdate)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProjectService_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type ProjectService_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - toUpdate project.Project
func (_e *ProjectService_Expecter) Update(ctx interface{}, toUpdate interface{}) *ProjectService_Update_Call {
	return &ProjectService_Update_Call{Call: _e.mock.On("Update", ctx, toUpdate)}
}

func (_c *ProjectService_Update_Call) Run(run func(ctx context.Context, toUpdate project.Project)) *ProjectService_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(project.Project))
	})
	return _c
}

func (_c *ProjectService_Update_Call) Return(_a0 project.Project, _a1 error) *ProjectService_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProjectService_Update_Call) RunAndReturn(run func(context.Context, project.Project) (project.Project, error)) *ProjectService_Update_Call {
	_c.Call.Return(run)
	return _c
}

// NewProjectService creates a new instance of ProjectService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProjectService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ProjectService {
	mock := &ProjectService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
