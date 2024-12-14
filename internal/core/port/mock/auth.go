// Code generated by MockGen. DO NOT EDIT.
// Source: auth.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	domain "github.com/MikeRez0/ypgophermart/internal/core/domain"
	port "github.com/MikeRez0/ypgophermart/internal/core/port"
	gomock "github.com/golang/mock/gomock"
)

// MockTokenService is a mock of TokenService interface.
type MockTokenService struct {
	ctrl     *gomock.Controller
	recorder *MockTokenServiceMockRecorder
}

// MockTokenServiceMockRecorder is the mock recorder for MockTokenService.
type MockTokenServiceMockRecorder struct {
	mock *MockTokenService
}

// NewMockTokenService creates a new mock instance.
func NewMockTokenService(ctrl *gomock.Controller) *MockTokenService {
	mock := &MockTokenService{ctrl: ctrl}
	mock.recorder = &MockTokenServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTokenService) EXPECT() *MockTokenServiceMockRecorder {
	return m.recorder
}

// CreateToken mocks base method.
func (m *MockTokenService) CreateToken(user *domain.User) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateToken", user)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateToken indicates an expected call of CreateToken.
func (mr *MockTokenServiceMockRecorder) CreateToken(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateToken", reflect.TypeOf((*MockTokenService)(nil).CreateToken), user)
}

// VerifyToken mocks base method.
func (m *MockTokenService) VerifyToken(token string) (*port.TokenPayload, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyToken", token)
	ret0, _ := ret[0].(*port.TokenPayload)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifyToken indicates an expected call of VerifyToken.
func (mr *MockTokenServiceMockRecorder) VerifyToken(token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyToken", reflect.TypeOf((*MockTokenService)(nil).VerifyToken), token)
}
