// Code generated by MockGen. DO NOT EDIT.
// Source: network/http.go

// Package mock is a generated GoMock package.
package mock

import (
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockHttp is a mock of Http interface.
type MockHttp struct {
	ctrl     *gomock.Controller
	recorder *MockHttpMockRecorder
}

// MockHttpMockRecorder is the mock recorder for MockHttp.
type MockHttpMockRecorder struct {
	mock *MockHttp
}

// NewMockHttp creates a new mock instance.
func NewMockHttp(ctrl *gomock.Controller) *MockHttp {
	mock := &MockHttp{ctrl: ctrl}
	mock.recorder = &MockHttpMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHttp) EXPECT() *MockHttpMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockHttp) Get(req *http.Request) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", req)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockHttpMockRecorder) Get(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockHttp)(nil).Get), req)
}
