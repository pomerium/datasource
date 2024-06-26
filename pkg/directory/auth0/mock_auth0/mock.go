// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/pomerium/datasource/pkg/directory/auth0 (interfaces: RoleManager,UserManager)
//
// Generated by this command:
//
//	mockgen -destination=mock_auth0/mock.go . RoleManager,UserManager
//

// Package mock_auth0 is a generated GoMock package.
package mock_auth0

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	management "gopkg.in/auth0.v5/management"
)

// MockRoleManager is a mock of RoleManager interface.
type MockRoleManager struct {
	ctrl     *gomock.Controller
	recorder *MockRoleManagerMockRecorder
}

// MockRoleManagerMockRecorder is the mock recorder for MockRoleManager.
type MockRoleManagerMockRecorder struct {
	mock *MockRoleManager
}

// NewMockRoleManager creates a new mock instance.
func NewMockRoleManager(ctrl *gomock.Controller) *MockRoleManager {
	mock := &MockRoleManager{ctrl: ctrl}
	mock.recorder = &MockRoleManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRoleManager) EXPECT() *MockRoleManagerMockRecorder {
	return m.recorder
}

// List mocks base method.
func (m *MockRoleManager) List(arg0 ...management.RequestOption) (*management.RoleList, error) {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "List", varargs...)
	ret0, _ := ret[0].(*management.RoleList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockRoleManagerMockRecorder) List(arg0 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockRoleManager)(nil).List), arg0...)
}

// Users mocks base method.
func (m *MockRoleManager) Users(arg0 string, arg1 ...management.RequestOption) (*management.UserList, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Users", varargs...)
	ret0, _ := ret[0].(*management.UserList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Users indicates an expected call of Users.
func (mr *MockRoleManagerMockRecorder) Users(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Users", reflect.TypeOf((*MockRoleManager)(nil).Users), varargs...)
}

// MockUserManager is a mock of UserManager interface.
type MockUserManager struct {
	ctrl     *gomock.Controller
	recorder *MockUserManagerMockRecorder
}

// MockUserManagerMockRecorder is the mock recorder for MockUserManager.
type MockUserManagerMockRecorder struct {
	mock *MockUserManager
}

// NewMockUserManager creates a new mock instance.
func NewMockUserManager(ctrl *gomock.Controller) *MockUserManager {
	mock := &MockUserManager{ctrl: ctrl}
	mock.recorder = &MockUserManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserManager) EXPECT() *MockUserManagerMockRecorder {
	return m.recorder
}

// Read mocks base method.
func (m *MockUserManager) Read(arg0 string, arg1 ...management.RequestOption) (*management.User, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Read", varargs...)
	ret0, _ := ret[0].(*management.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockUserManagerMockRecorder) Read(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockUserManager)(nil).Read), varargs...)
}

// Roles mocks base method.
func (m *MockUserManager) Roles(arg0 string, arg1 ...management.RequestOption) (*management.RoleList, error) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Roles", varargs...)
	ret0, _ := ret[0].(*management.RoleList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Roles indicates an expected call of Roles.
func (mr *MockUserManagerMockRecorder) Roles(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Roles", reflect.TypeOf((*MockUserManager)(nil).Roles), varargs...)
}
