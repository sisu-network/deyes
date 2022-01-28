// Code generated by MockGen. DO NOT EDIT.
// Source: database/db.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/sisu-network/deyes/types"
)

// MockDatabase is a mock of Database interface.
type MockDatabase struct {
	ctrl     *gomock.Controller
	recorder *MockDatabaseMockRecorder
}

// MockDatabaseMockRecorder is the mock recorder for MockDatabase.
type MockDatabaseMockRecorder struct {
	mock *MockDatabase
}

// NewMockDatabase creates a new mock instance.
func NewMockDatabase(ctrl *gomock.Controller) *MockDatabase {
	mock := &MockDatabase{ctrl: ctrl}
	mock.recorder = &MockDatabaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDatabase) EXPECT() *MockDatabaseMockRecorder {
	return m.recorder
}

// Init mocks base method.
func (m *MockDatabase) Init() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Init")
	ret0, _ := ret[0].(error)
	return ret0
}

// Init indicates an expected call of Init.
func (mr *MockDatabaseMockRecorder) Init() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockDatabase)(nil).Init))
}

// LoadPrices mocks base method.
func (m *MockDatabase) LoadPrices() types.TokenPrices {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadPrices")
	ret0, _ := ret[0].(types.TokenPrices)
	return ret0
}

// LoadPrices indicates an expected call of LoadPrices.
func (mr *MockDatabaseMockRecorder) LoadPrices() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadPrices", reflect.TypeOf((*MockDatabase)(nil).LoadPrices))
}

// LoadWatchAddresses mocks base method.
func (m *MockDatabase) LoadWatchAddresses(chain string) []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadWatchAddresses", chain)
	ret0, _ := ret[0].([]string)
	return ret0
}

// LoadWatchAddresses indicates an expected call of LoadWatchAddresses.
func (mr *MockDatabaseMockRecorder) LoadWatchAddresses(chain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadWatchAddresses", reflect.TypeOf((*MockDatabase)(nil).LoadWatchAddresses), chain)
}

// SaveTokenPrices mocks base method.
func (m *MockDatabase) SaveTokenPrices(tokenPrices types.TokenPrices) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SaveTokenPrices", tokenPrices)
}

// SaveTokenPrices indicates an expected call of SaveTokenPrices.
func (mr *MockDatabaseMockRecorder) SaveTokenPrices(tokenPrices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveTokenPrices", reflect.TypeOf((*MockDatabase)(nil).SaveTokenPrices), tokenPrices)
}

// SaveTxs mocks base method.
func (m *MockDatabase) SaveTxs(chain string, blockHeight int64, txs *types.Txs) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SaveTxs", chain, blockHeight, txs)
}

// SaveTxs indicates an expected call of SaveTxs.
func (mr *MockDatabaseMockRecorder) SaveTxs(chain, blockHeight, txs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveTxs", reflect.TypeOf((*MockDatabase)(nil).SaveTxs), chain, blockHeight, txs)
}

// SaveWatchAddress mocks base method.
func (m *MockDatabase) SaveWatchAddress(chain, address string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SaveWatchAddress", chain, address)
}

// SaveWatchAddress indicates an expected call of SaveWatchAddress.
func (mr *MockDatabaseMockRecorder) SaveWatchAddress(chain, address interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveWatchAddress", reflect.TypeOf((*MockDatabase)(nil).SaveWatchAddress), chain, address)
}
