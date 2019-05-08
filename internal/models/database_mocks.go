// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.com/teserakt/c2/internal/models (interfaces: Database)

// Package models is a generated GoMock package.
package models

import (
	gomock "github.com/golang/mock/gomock"
	gorm "github.com/jinzhu/gorm"
	reflect "reflect"
)

// MockDatabase is a mock of Database interface
type MockDatabase struct {
	ctrl     *gomock.Controller
	recorder *MockDatabaseMockRecorder
}

// MockDatabaseMockRecorder is the mock recorder for MockDatabase
type MockDatabaseMockRecorder struct {
	mock *MockDatabase
}

// NewMockDatabase creates a new mock instance
func NewMockDatabase(ctrl *gomock.Controller) *MockDatabase {
	mock := &MockDatabase{ctrl: ctrl}
	mock.recorder = &MockDatabaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDatabase) EXPECT() *MockDatabaseMockRecorder {
	return m.recorder
}

// Close mocks base method
func (m *MockDatabase) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockDatabaseMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockDatabase)(nil).Close))
}

// Connection mocks base method
func (m *MockDatabase) Connection() *gorm.DB {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connection")
	ret0, _ := ret[0].(*gorm.DB)
	return ret0
}

// Connection indicates an expected call of Connection
func (mr *MockDatabaseMockRecorder) Connection() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connection", reflect.TypeOf((*MockDatabase)(nil).Connection))
}

// CountIDKeys mocks base method
func (m *MockDatabase) CountIDKeys() (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountIDKeys")
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CountIDKeys indicates an expected call of CountIDKeys
func (mr *MockDatabaseMockRecorder) CountIDKeys() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountIDKeys", reflect.TypeOf((*MockDatabase)(nil).CountIDKeys))
}

// CountIDsForTopic mocks base method
func (m *MockDatabase) CountIDsForTopic(arg0 string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountIDsForTopic", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CountIDsForTopic indicates an expected call of CountIDsForTopic
func (mr *MockDatabaseMockRecorder) CountIDsForTopic(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountIDsForTopic", reflect.TypeOf((*MockDatabase)(nil).CountIDsForTopic), arg0)
}

// CountTopicKeys mocks base method
func (m *MockDatabase) CountTopicKeys() (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountTopicKeys")
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CountTopicKeys indicates an expected call of CountTopicKeys
func (mr *MockDatabaseMockRecorder) CountTopicKeys() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountTopicKeys", reflect.TypeOf((*MockDatabase)(nil).CountTopicKeys))
}

// CountTopicsForID mocks base method
func (m *MockDatabase) CountTopicsForID(arg0 []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountTopicsForID", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CountTopicsForID indicates an expected call of CountTopicsForID
func (mr *MockDatabaseMockRecorder) CountTopicsForID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountTopicsForID", reflect.TypeOf((*MockDatabase)(nil).CountTopicsForID), arg0)
}

// DeleteIDKey mocks base method
func (m *MockDatabase) DeleteIDKey(arg0 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteIDKey", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteIDKey indicates an expected call of DeleteIDKey
func (mr *MockDatabaseMockRecorder) DeleteIDKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteIDKey", reflect.TypeOf((*MockDatabase)(nil).DeleteIDKey), arg0)
}

// DeleteTopicKey mocks base method
func (m *MockDatabase) DeleteTopicKey(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteTopicKey", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteTopicKey indicates an expected call of DeleteTopicKey
func (mr *MockDatabaseMockRecorder) DeleteTopicKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteTopicKey", reflect.TypeOf((*MockDatabase)(nil).DeleteTopicKey), arg0)
}

// GetAllIDKeys mocks base method
func (m *MockDatabase) GetAllIDKeys() ([]IDKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllIDKeys")
	ret0, _ := ret[0].([]IDKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllIDKeys indicates an expected call of GetAllIDKeys
func (mr *MockDatabaseMockRecorder) GetAllIDKeys() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllIDKeys", reflect.TypeOf((*MockDatabase)(nil).GetAllIDKeys))
}

// GetAllTopics mocks base method
func (m *MockDatabase) GetAllTopics() ([]TopicKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllTopics")
	ret0, _ := ret[0].([]TopicKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllTopics indicates an expected call of GetAllTopics
func (mr *MockDatabaseMockRecorder) GetAllTopics() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllTopics", reflect.TypeOf((*MockDatabase)(nil).GetAllTopics))
}

// GetIDKey mocks base method
func (m *MockDatabase) GetIDKey(arg0 []byte) (IDKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIDKey", arg0)
	ret0, _ := ret[0].(IDKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetIDKey indicates an expected call of GetIDKey
func (mr *MockDatabaseMockRecorder) GetIDKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIDKey", reflect.TypeOf((*MockDatabase)(nil).GetIDKey), arg0)
}

// GetIdsforTopic mocks base method
func (m *MockDatabase) GetIdsforTopic(arg0 string, arg1, arg2 int) ([]IDKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIdsforTopic", arg0, arg1, arg2)
	ret0, _ := ret[0].([]IDKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetIdsforTopic indicates an expected call of GetIdsforTopic
func (mr *MockDatabaseMockRecorder) GetIdsforTopic(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIdsforTopic", reflect.TypeOf((*MockDatabase)(nil).GetIdsforTopic), arg0, arg1, arg2)
}

// GetTopicKey mocks base method
func (m *MockDatabase) GetTopicKey(arg0 string) (TopicKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTopicKey", arg0)
	ret0, _ := ret[0].(TopicKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTopicKey indicates an expected call of GetTopicKey
func (mr *MockDatabaseMockRecorder) GetTopicKey(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTopicKey", reflect.TypeOf((*MockDatabase)(nil).GetTopicKey), arg0)
}

// GetTopicsForID mocks base method
func (m *MockDatabase) GetTopicsForID(arg0 []byte, arg1, arg2 int) ([]TopicKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTopicsForID", arg0, arg1, arg2)
	ret0, _ := ret[0].([]TopicKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTopicsForID indicates an expected call of GetTopicsForID
func (mr *MockDatabaseMockRecorder) GetTopicsForID(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTopicsForID", reflect.TypeOf((*MockDatabase)(nil).GetTopicsForID), arg0, arg1, arg2)
}

// InsertIDKey mocks base method
func (m *MockDatabase) InsertIDKey(arg0, arg1 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertIDKey", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertIDKey indicates an expected call of InsertIDKey
func (mr *MockDatabaseMockRecorder) InsertIDKey(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertIDKey", reflect.TypeOf((*MockDatabase)(nil).InsertIDKey), arg0, arg1)
}

// InsertTopicKey mocks base method
func (m *MockDatabase) InsertTopicKey(arg0 string, arg1 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertTopicKey", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertTopicKey indicates an expected call of InsertTopicKey
func (mr *MockDatabaseMockRecorder) InsertTopicKey(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertTopicKey", reflect.TypeOf((*MockDatabase)(nil).InsertTopicKey), arg0, arg1)
}

// LinkIDTopic mocks base method
func (m *MockDatabase) LinkIDTopic(arg0 IDKey, arg1 TopicKey) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LinkIDTopic", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// LinkIDTopic indicates an expected call of LinkIDTopic
func (mr *MockDatabaseMockRecorder) LinkIDTopic(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LinkIDTopic", reflect.TypeOf((*MockDatabase)(nil).LinkIDTopic), arg0, arg1)
}

// Migrate mocks base method
func (m *MockDatabase) Migrate() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Migrate")
	ret0, _ := ret[0].(error)
	return ret0
}

// Migrate indicates an expected call of Migrate
func (mr *MockDatabaseMockRecorder) Migrate() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Migrate", reflect.TypeOf((*MockDatabase)(nil).Migrate))
}

// UnlinkIDTopic mocks base method
func (m *MockDatabase) UnlinkIDTopic(arg0 IDKey, arg1 TopicKey) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnlinkIDTopic", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UnlinkIDTopic indicates an expected call of UnlinkIDTopic
func (mr *MockDatabaseMockRecorder) UnlinkIDTopic(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnlinkIDTopic", reflect.TypeOf((*MockDatabase)(nil).UnlinkIDTopic), arg0, arg1)
}
