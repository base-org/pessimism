// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/base-org/pessimism/internal/etl/pipeline (interfaces: Manager)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	core "github.com/base-org/pessimism/internal/core"
	gomock "github.com/golang/mock/gomock"
)

// EtlManager is a mock of Manager interface.
type EtlManager struct {
	ctrl     *gomock.Controller
	recorder *EtlManagerMockRecorder
}

// EtlManagerMockRecorder is the mock recorder for EtlManager.
type EtlManagerMockRecorder struct {
	mock *EtlManager
}

// NewEtlManager creates a new mock instance.
func NewEtlManager(ctrl *gomock.Controller) *EtlManager {
	mock := &EtlManager{ctrl: ctrl}
	mock.recorder = &EtlManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *EtlManager) EXPECT() *EtlManagerMockRecorder {
	return m.recorder
}

// CreateDataPipeline mocks base method.
func (m *EtlManager) CreateDataPipeline(arg0 *core.PipelineConfig) (core.PUUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateDataPipeline", arg0)
	ret0, _ := ret[0].(core.PUUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateDataPipeline indicates an expected call of CreateDataPipeline.
func (mr *EtlManagerMockRecorder) CreateDataPipeline(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDataPipeline", reflect.TypeOf((*EtlManager)(nil).CreateDataPipeline), arg0)
}

// EventLoop mocks base method.
func (m *EtlManager) EventLoop(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EventLoop", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// EventLoop indicates an expected call of EventLoop.
func (mr *EtlManagerMockRecorder) EventLoop(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EventLoop", reflect.TypeOf((*EtlManager)(nil).EventLoop), arg0)
}

// GetRegister mocks base method.
func (m *EtlManager) GetRegister(arg0 core.RegisterType) (*core.DataRegister, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRegister", arg0)
	ret0, _ := ret[0].(*core.DataRegister)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRegister indicates an expected call of GetRegister.
func (mr *EtlManagerMockRecorder) GetRegister(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRegister", reflect.TypeOf((*EtlManager)(nil).GetRegister), arg0)
}

// RunPipeline mocks base method.
func (m *EtlManager) RunPipeline(arg0 core.PUUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RunPipeline", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RunPipeline indicates an expected call of RunPipeline.
func (mr *EtlManagerMockRecorder) RunPipeline(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunPipeline", reflect.TypeOf((*EtlManager)(nil).RunPipeline), arg0)
}
