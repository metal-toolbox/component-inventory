// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/api/client/client.go
//
// Generated by this command:
//
//	mockgen -package=client -source=pkg/api/client/client.go -destination=pkg/api/client/mock/mockclient.go
//

// Package client is a generated GoMock package.
package client

import (
	context "context"
	reflect "reflect"

	common "github.com/bmc-toolbox/common"
	client "github.com/metal-toolbox/component-inventory/pkg/api/client"
	gomock "go.uber.org/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// GetServerComponents mocks base method.
func (m *MockClient) GetServerComponents(arg0 context.Context, arg1 string) (client.ServerComponents, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServerComponents", arg0, arg1)
	ret0, _ := ret[0].(client.ServerComponents)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetServerComponents indicates an expected call of GetServerComponents.
func (mr *MockClientMockRecorder) GetServerComponents(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServerComponents", reflect.TypeOf((*MockClient)(nil).GetServerComponents), arg0, arg1)
}

// UpdateInbandInventory mocks base method.
func (m *MockClient) UpdateInbandInventory(arg0 context.Context, arg1 string, arg2 *common.Device) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateInbandInventory", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateInbandInventory indicates an expected call of UpdateInbandInventory.
func (mr *MockClientMockRecorder) UpdateInbandInventory(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateInbandInventory", reflect.TypeOf((*MockClient)(nil).UpdateInbandInventory), arg0, arg1, arg2)
}

// UpdateOutOfbandInventory mocks base method.
func (m *MockClient) UpdateOutOfbandInventory(arg0 context.Context, arg1 string, arg2 *common.Device) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOutOfbandInventory", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateOutOfbandInventory indicates an expected call of UpdateOutOfbandInventory.
func (mr *MockClientMockRecorder) UpdateOutOfbandInventory(arg0, arg1, arg2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOutOfbandInventory", reflect.TypeOf((*MockClient)(nil).UpdateOutOfbandInventory), arg0, arg1, arg2)
}

// Version mocks base method.
func (m *MockClient) Version(arg0 context.Context) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Version", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Version indicates an expected call of Version.
func (mr *MockClientMockRecorder) Version(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Version", reflect.TypeOf((*MockClient)(nil).Version), arg0)
}
