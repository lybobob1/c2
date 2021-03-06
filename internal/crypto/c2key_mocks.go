// Copyright 2020 Teserakt AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/teserakt-io/c2/internal/crypto (interfaces: C2KeyRotationTx)

// Package crypto is a generated GoMock package.
package crypto

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockC2KeyRotationTx is a mock of C2KeyRotationTx interface
type MockC2KeyRotationTx struct {
	ctrl     *gomock.Controller
	recorder *MockC2KeyRotationTxMockRecorder
}

// MockC2KeyRotationTxMockRecorder is the mock recorder for MockC2KeyRotationTx
type MockC2KeyRotationTxMockRecorder struct {
	mock *MockC2KeyRotationTx
}

// NewMockC2KeyRotationTx creates a new mock instance
func NewMockC2KeyRotationTx(ctrl *gomock.Controller) *MockC2KeyRotationTx {
	mock := &MockC2KeyRotationTx{ctrl: ctrl}
	mock.recorder = &MockC2KeyRotationTxMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockC2KeyRotationTx) EXPECT() *MockC2KeyRotationTxMockRecorder {
	return m.recorder
}

// Commit mocks base method
func (m *MockC2KeyRotationTx) Commit() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit")
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit
func (mr *MockC2KeyRotationTxMockRecorder) Commit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockC2KeyRotationTx)(nil).Commit))
}

// GetNewPublicKey mocks base method
func (m *MockC2KeyRotationTx) GetNewPublicKey() []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNewPublicKey")
	ret0, _ := ret[0].([]byte)
	return ret0
}

// GetNewPublicKey indicates an expected call of GetNewPublicKey
func (mr *MockC2KeyRotationTxMockRecorder) GetNewPublicKey() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNewPublicKey", reflect.TypeOf((*MockC2KeyRotationTx)(nil).GetNewPublicKey))
}

// Rollback mocks base method
func (m *MockC2KeyRotationTx) Rollback() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rollback")
	ret0, _ := ret[0].(error)
	return ret0
}

// Rollback indicates an expected call of Rollback
func (mr *MockC2KeyRotationTxMockRecorder) Rollback() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", reflect.TypeOf((*MockC2KeyRotationTx)(nil).Rollback))
}
