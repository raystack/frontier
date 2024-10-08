// Code generated by mockery v2.40.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// PreferencesService is an autogenerated mock type for the PreferencesService type
type PreferencesService struct {
	mock.Mock
}

type PreferencesService_Expecter struct {
	mock *mock.Mock
}

func (_m *PreferencesService) EXPECT() *PreferencesService_Expecter {
	return &PreferencesService_Expecter{mock: &_m.Mock}
}

// LoadPlatformPreferences provides a mock function with given fields: ctx
func (_m *PreferencesService) LoadPlatformPreferences(ctx context.Context) (map[string]string, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for LoadPlatformPreferences")
	}

	var r0 map[string]string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (map[string]string, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) map[string]string); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PreferencesService_LoadPlatformPreferences_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LoadPlatformPreferences'
type PreferencesService_LoadPlatformPreferences_Call struct {
	*mock.Call
}

// LoadPlatformPreferences is a helper method to define mock.On call
//   - ctx context.Context
func (_e *PreferencesService_Expecter) LoadPlatformPreferences(ctx interface{}) *PreferencesService_LoadPlatformPreferences_Call {
	return &PreferencesService_LoadPlatformPreferences_Call{Call: _e.mock.On("LoadPlatformPreferences", ctx)}
}

func (_c *PreferencesService_LoadPlatformPreferences_Call) Run(run func(ctx context.Context)) *PreferencesService_LoadPlatformPreferences_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *PreferencesService_LoadPlatformPreferences_Call) Return(_a0 map[string]string, _a1 error) *PreferencesService_LoadPlatformPreferences_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *PreferencesService_LoadPlatformPreferences_Call) RunAndReturn(run func(context.Context) (map[string]string, error)) *PreferencesService_LoadPlatformPreferences_Call {
	_c.Call.Return(run)
	return _c
}

// NewPreferencesService creates a new instance of PreferencesService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPreferencesService(t interface {
	mock.TestingT
	Cleanup(func())
}) *PreferencesService {
	mock := &PreferencesService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
