// Code generated by mockery v2.45.0. DO NOT EDIT.

package mocks

import (
	mail "gopkg.in/mail.v2"

	mock "github.com/stretchr/testify/mock"
)

// Dialer is an autogenerated mock type for the Dialer type
type Dialer struct {
	mock.Mock
}

type Dialer_Expecter struct {
	mock *mock.Mock
}

func (_m *Dialer) EXPECT() *Dialer_Expecter {
	return &Dialer_Expecter{mock: &_m.Mock}
}

// DialAndSend provides a mock function with given fields: m
func (_m *Dialer) DialAndSend(m *mail.Message) error {
	ret := _m.Called(m)

	if len(ret) == 0 {
		panic("no return value specified for DialAndSend")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*mail.Message) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Dialer_DialAndSend_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DialAndSend'
type Dialer_DialAndSend_Call struct {
	*mock.Call
}

// DialAndSend is a helper method to define mock.On call
//   - m *mail.Message
func (_e *Dialer_Expecter) DialAndSend(m interface{}) *Dialer_DialAndSend_Call {
	return &Dialer_DialAndSend_Call{Call: _e.mock.On("DialAndSend", m)}
}

func (_c *Dialer_DialAndSend_Call) Run(run func(m *mail.Message)) *Dialer_DialAndSend_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*mail.Message))
	})
	return _c
}

func (_c *Dialer_DialAndSend_Call) Return(_a0 error) *Dialer_DialAndSend_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Dialer_DialAndSend_Call) RunAndReturn(run func(*mail.Message) error) *Dialer_DialAndSend_Call {
	_c.Call.Return(run)
	return _c
}

// FromHeader provides a mock function with given fields:
func (_m *Dialer) FromHeader() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for FromHeader")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Dialer_FromHeader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FromHeader'
type Dialer_FromHeader_Call struct {
	*mock.Call
}

// FromHeader is a helper method to define mock.On call
func (_e *Dialer_Expecter) FromHeader() *Dialer_FromHeader_Call {
	return &Dialer_FromHeader_Call{Call: _e.mock.On("FromHeader")}
}

func (_c *Dialer_FromHeader_Call) Run(run func()) *Dialer_FromHeader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Dialer_FromHeader_Call) Return(_a0 string) *Dialer_FromHeader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Dialer_FromHeader_Call) RunAndReturn(run func() string) *Dialer_FromHeader_Call {
	_c.Call.Return(run)
	return _c
}

// NewDialer creates a new instance of Dialer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDialer(t interface {
	mock.TestingT
	Cleanup(func())
}) *Dialer {
	mock := &Dialer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
