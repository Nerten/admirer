// Code generated by MockGen. DO NOT EDIT.
// Source: authentication.go

// Package authentication is a generated GoMock package.
package authentication

import (
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockCallbackProvider is a mock of CallbackProvider interface.
type MockCallbackProvider struct {
	ctrl     *gomock.Controller
	recorder *MockCallbackProviderMockRecorder
}

// MockCallbackProviderMockRecorder is the mock recorder for MockCallbackProvider.
type MockCallbackProviderMockRecorder struct {
	mock *MockCallbackProvider
}

// NewMockCallbackProvider creates a new mock instance.
func NewMockCallbackProvider(ctrl *gomock.Controller) *MockCallbackProvider {
	mock := &MockCallbackProvider{ctrl: ctrl}
	mock.recorder = &MockCallbackProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCallbackProvider) EXPECT() *MockCallbackProviderMockRecorder {
	return m.recorder
}

// ReadCode mocks base method.
func (m *MockCallbackProvider) ReadCode(key string, writer io.Writer) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadCode", key, writer)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadCode indicates an expected call of ReadCode.
func (mr *MockCallbackProviderMockRecorder) ReadCode(key, writer interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadCode", reflect.TypeOf((*MockCallbackProvider)(nil).ReadCode), key, writer)
}
