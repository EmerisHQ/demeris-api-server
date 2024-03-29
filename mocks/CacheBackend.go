// Code generated by mockery v2.12.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	testing "testing"

	time "time"
)

// CacheBackend is an autogenerated mock type for the CacheBackend type
type CacheBackend struct {
	mock.Mock
}

type CacheBackend_Expecter struct {
	mock *mock.Mock
}

func (_m *CacheBackend) EXPECT() *CacheBackend_Expecter {
	return &CacheBackend_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: ctx, key
func (_m *CacheBackend) Get(ctx context.Context, key string) (string, error) {
	ret := _m.Called(ctx, key)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CacheBackend_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type CacheBackend_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//  - ctx context.Context
//  - key string
func (_e *CacheBackend_Expecter) Get(ctx interface{}, key interface{}) *CacheBackend_Get_Call {
	return &CacheBackend_Get_Call{Call: _e.mock.On("Get", ctx, key)}
}

func (_c *CacheBackend_Get_Call) Run(run func(ctx context.Context, key string)) *CacheBackend_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CacheBackend_Get_Call) Return(_a0 string, _a1 error) *CacheBackend_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Set provides a mock function with given fields: ctx, key, value, expiration
func (_m *CacheBackend) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	ret := _m.Called(ctx, key, value, expiration)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, time.Duration) error); ok {
		r0 = rf(ctx, key, value, expiration)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CacheBackend_Set_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Set'
type CacheBackend_Set_Call struct {
	*mock.Call
}

// Set is a helper method to define mock.On call
//  - ctx context.Context
//  - key string
//  - value string
//  - expiration time.Duration
func (_e *CacheBackend_Expecter) Set(ctx interface{}, key interface{}, value interface{}, expiration interface{}) *CacheBackend_Set_Call {
	return &CacheBackend_Set_Call{Call: _e.mock.On("Set", ctx, key, value, expiration)}
}

func (_c *CacheBackend_Set_Call) Run(run func(ctx context.Context, key string, value string, expiration time.Duration)) *CacheBackend_Set_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(time.Duration))
	})
	return _c
}

func (_c *CacheBackend_Set_Call) Return(_a0 error) *CacheBackend_Set_Call {
	_c.Call.Return(_a0)
	return _c
}

// NewCacheBackend creates a new instance of CacheBackend. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewCacheBackend(t testing.TB) *CacheBackend {
	mock := &CacheBackend{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
