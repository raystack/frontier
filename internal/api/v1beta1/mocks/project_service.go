// Code generated by mockery v2.20.2. DO NOT EDIT.

package mocks

import (
	context "context"

	project "github.com/odpf/shield/core/project"
	mock "github.com/stretchr/testify/mock"

	user "github.com/odpf/shield/core/user"
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

// Create provides a mock function with given fields: ctx, prj
func (_m *ProjectService) Create(ctx context.Context, prj project.Project) (project.Project, error) {
	ret := _m.Called(ctx, prj)

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

// Get provides a mock function with given fields: ctx, idOrSlugd
func (_m *ProjectService) Get(ctx context.Context, idOrSlugd string) (project.Project, error) {
	ret := _m.Called(ctx, idOrSlugd)

	var r0 project.Project
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (project.Project, error)); ok {
		return rf(ctx, idOrSlugd)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) project.Project); ok {
		r0 = rf(ctx, idOrSlugd)
	} else {
		r0 = ret.Get(0).(project.Project)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, idOrSlugd)
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
//   - idOrSlugd string
func (_e *ProjectService_Expecter) Get(ctx interface{}, idOrSlugd interface{}) *ProjectService_Get_Call {
	return &ProjectService_Get_Call{Call: _e.mock.On("Get", ctx, idOrSlugd)}
}

func (_c *ProjectService_Get_Call) Run(run func(ctx context.Context, idOrSlugd string)) *ProjectService_Get_Call {
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

// List provides a mock function with given fields: ctx
func (_m *ProjectService) List(ctx context.Context) ([]project.Project, error) {
	ret := _m.Called(ctx)

	var r0 []project.Project
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]project.Project, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []project.Project); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]project.Project)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
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
func (_e *ProjectService_Expecter) List(ctx interface{}) *ProjectService_List_Call {
	return &ProjectService_List_Call{Call: _e.mock.On("List", ctx)}
}

func (_c *ProjectService_List_Call) Run(run func(ctx context.Context)) *ProjectService_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *ProjectService_List_Call) Return(_a0 []project.Project, _a1 error) *ProjectService_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProjectService_List_Call) RunAndReturn(run func(context.Context) ([]project.Project, error)) *ProjectService_List_Call {
	_c.Call.Return(run)
	return _c
}

// ListByOrganization provides a mock function with given fields: ctx, organizationID
func (_m *ProjectService) ListByOrganization(ctx context.Context, organizationID string) ([]project.Project, error) {
	ret := _m.Called(ctx, organizationID)

	var r0 []project.Project
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]project.Project, error)); ok {
		return rf(ctx, organizationID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []project.Project); ok {
		r0 = rf(ctx, organizationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]project.Project)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, organizationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProjectService_ListByOrganization_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListByOrganization'
type ProjectService_ListByOrganization_Call struct {
	*mock.Call
}

// ListByOrganization is a helper method to define mock.On call
//   - ctx context.Context
//   - organizationID string
func (_e *ProjectService_Expecter) ListByOrganization(ctx interface{}, organizationID interface{}) *ProjectService_ListByOrganization_Call {
	return &ProjectService_ListByOrganization_Call{Call: _e.mock.On("ListByOrganization", ctx, organizationID)}
}

func (_c *ProjectService_ListByOrganization_Call) Run(run func(ctx context.Context, organizationID string)) *ProjectService_ListByOrganization_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *ProjectService_ListByOrganization_Call) Return(_a0 []project.Project, _a1 error) *ProjectService_ListByOrganization_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ProjectService_ListByOrganization_Call) RunAndReturn(run func(context.Context, string) ([]project.Project, error)) *ProjectService_ListByOrganization_Call {
	_c.Call.Return(run)
	return _c
}

// ListUsers provides a mock function with given fields: ctx, id, permissionFilter
func (_m *ProjectService) ListUsers(ctx context.Context, id string, permissionFilter string) ([]user.User, error) {
	ret := _m.Called(ctx, id, permissionFilter)

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

type mockConstructorTestingTNewProjectService interface {
	mock.TestingT
	Cleanup(func())
}

// NewProjectService creates a new instance of ProjectService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewProjectService(t mockConstructorTestingTNewProjectService) *ProjectService {
	mock := &ProjectService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
