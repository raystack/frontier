// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	context "context"

	session "github.com/raystack/frontier/core/authenticate/session"
	mock "github.com/stretchr/testify/mock"

	time "time"

	uuid "github.com/google/uuid"
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

// Delete provides a mock function with given fields: ctx, id
func (_m *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type Repository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - id uuid.UUID
func (_e *Repository_Expecter) Delete(ctx interface{}, id interface{}) *Repository_Delete_Call {
	return &Repository_Delete_Call{Call: _e.mock.On("Delete", ctx, id)}
}

func (_c *Repository_Delete_Call) Run(run func(ctx context.Context, id uuid.UUID)) *Repository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uuid.UUID))
	})
	return _c
}

func (_c *Repository_Delete_Call) Return(_a0 error) *Repository_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repository_Delete_Call) RunAndReturn(run func(context.Context, uuid.UUID) error) *Repository_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteExpiredSessions provides a mock function with given fields: ctx
func (_m *Repository) DeleteExpiredSessions(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for DeleteExpiredSessions")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repository_DeleteExpiredSessions_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteExpiredSessions'
type Repository_DeleteExpiredSessions_Call struct {
	*mock.Call
}

// DeleteExpiredSessions is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Repository_Expecter) DeleteExpiredSessions(ctx interface{}) *Repository_DeleteExpiredSessions_Call {
	return &Repository_DeleteExpiredSessions_Call{Call: _e.mock.On("DeleteExpiredSessions", ctx)}
}

func (_c *Repository_DeleteExpiredSessions_Call) Run(run func(ctx context.Context)) *Repository_DeleteExpiredSessions_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Repository_DeleteExpiredSessions_Call) Return(_a0 error) *Repository_DeleteExpiredSessions_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repository_DeleteExpiredSessions_Call) RunAndReturn(run func(context.Context) error) *Repository_DeleteExpiredSessions_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, id
func (_m *Repository) Get(ctx context.Context, id uuid.UUID) (*session.Session, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *session.Session
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (*session.Session, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *session.Session); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*session.Session)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Repository_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type Repository_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - id uuid.UUID
func (_e *Repository_Expecter) Get(ctx interface{}, id interface{}) *Repository_Get_Call {
	return &Repository_Get_Call{Call: _e.mock.On("Get", ctx, id)}
}

func (_c *Repository_Get_Call) Run(run func(ctx context.Context, id uuid.UUID)) *Repository_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uuid.UUID))
	})
	return _c
}

func (_c *Repository_Get_Call) Return(_a0 *session.Session, _a1 error) *Repository_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Repository_Get_Call) RunAndReturn(run func(context.Context, uuid.UUID) (*session.Session, error)) *Repository_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Set provides a mock function with given fields: ctx, _a1
func (_m *Repository) Set(ctx context.Context, _a1 *session.Session) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Set")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *session.Session) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repository_Set_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Set'
type Repository_Set_Call struct {
	*mock.Call
}

// Set is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 *session.Session
func (_e *Repository_Expecter) Set(ctx interface{}, _a1 interface{}) *Repository_Set_Call {
	return &Repository_Set_Call{Call: _e.mock.On("Set", ctx, _a1)}
}

func (_c *Repository_Set_Call) Run(run func(ctx context.Context, _a1 *session.Session)) *Repository_Set_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*session.Session))
	})
	return _c
}

func (_c *Repository_Set_Call) Return(_a0 error) *Repository_Set_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repository_Set_Call) RunAndReturn(run func(context.Context, *session.Session) error) *Repository_Set_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateValidity provides a mock function with given fields: ctx, id, validity
func (_m *Repository) UpdateValidity(ctx context.Context, id uuid.UUID, validity time.Duration) error {
	ret := _m.Called(ctx, id, validity)

	if len(ret) == 0 {
		panic("no return value specified for UpdateValidity")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, time.Duration) error); ok {
		r0 = rf(ctx, id, validity)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repository_UpdateValidity_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateValidity'
type Repository_UpdateValidity_Call struct {
	*mock.Call
}

// UpdateValidity is a helper method to define mock.On call
//   - ctx context.Context
//   - id uuid.UUID
//   - validity time.Duration
func (_e *Repository_Expecter) UpdateValidity(ctx interface{}, id interface{}, validity interface{}) *Repository_UpdateValidity_Call {
	return &Repository_UpdateValidity_Call{Call: _e.mock.On("UpdateValidity", ctx, id, validity)}
}

func (_c *Repository_UpdateValidity_Call) Run(run func(ctx context.Context, id uuid.UUID, validity time.Duration)) *Repository_UpdateValidity_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uuid.UUID), args[2].(time.Duration))
	})
	return _c
}

func (_c *Repository_UpdateValidity_Call) Return(_a0 error) *Repository_UpdateValidity_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repository_UpdateValidity_Call) RunAndReturn(run func(context.Context, uuid.UUID, time.Duration) error) *Repository_UpdateValidity_Call {
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
