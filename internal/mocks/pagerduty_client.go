// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/base-org/pessimism/internal/client (interfaces: PagerDutyClient)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	client "github.com/base-org/pessimism/internal/client"
	gomock "github.com/golang/mock/gomock"
)

// MockPagerDutyClient is a mock of PagerDutyClient interface.
type MockPagerDutyClient struct {
	ctrl     *gomock.Controller
	recorder *MockPagerDutyClientMockRecorder
}

// MockPagerDutyClientMockRecorder is the mock recorder for MockPagerDutyClient.
type MockPagerDutyClientMockRecorder struct {
	mock *MockPagerDutyClient
}

// NewMockPagerDutyClient creates a new mock instance.
func NewMockPagerDutyClient(ctrl *gomock.Controller) *MockPagerDutyClient {
	mock := &MockPagerDutyClient{ctrl: ctrl}
	mock.recorder = &MockPagerDutyClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPagerDutyClient) EXPECT() *MockPagerDutyClientMockRecorder {
	return m.recorder
}

// GetName mocks base method.
func (m *MockPagerDutyClient) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName.
func (mr *MockPagerDutyClientMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockPagerDutyClient)(nil).GetName))
}

// PostEvent mocks base method.
func (m *MockPagerDutyClient) PostEvent(arg0 context.Context, arg1 *client.AlertEventTrigger) (*client.AlertAPIResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostEvent", arg0, arg1)
	ret0, _ := ret[0].(*client.AlertAPIResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PostEvent indicates an expected call of PostEvent.
func (mr *MockPagerDutyClientMockRecorder) PostEvent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostEvent", reflect.TypeOf((*MockPagerDutyClient)(nil).PostEvent), arg0, arg1)
}
