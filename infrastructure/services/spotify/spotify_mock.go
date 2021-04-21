// Code generated by MockGen. DO NOT EDIT.
// Source: spotify.go

// Package spotify is a generated GoMock package.
package spotify

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	spotify "github.com/zmb3/spotify"
	oauth2 "golang.org/x/oauth2"
)

// MockAuthenticator is a mock of Authenticator interface.
type MockAuthenticator struct {
	ctrl     *gomock.Controller
	recorder *MockAuthenticatorMockRecorder
}

// MockAuthenticatorMockRecorder is the mock recorder for MockAuthenticator.
type MockAuthenticatorMockRecorder struct {
	mock *MockAuthenticator
}

// NewMockAuthenticator creates a new mock instance.
func NewMockAuthenticator(ctrl *gomock.Controller) *MockAuthenticator {
	mock := &MockAuthenticator{ctrl: ctrl}
	mock.recorder = &MockAuthenticatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAuthenticator) EXPECT() *MockAuthenticatorMockRecorder {
	return m.recorder
}

// AuthURLWithOpts mocks base method.
func (m *MockAuthenticator) AuthURLWithOpts(state string, opts ...oauth2.AuthCodeOption) string {
	m.ctrl.T.Helper()
	varargs := []interface{}{state}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "AuthURLWithOpts", varargs...)
	ret0, _ := ret[0].(string)
	return ret0
}

// AuthURLWithOpts indicates an expected call of AuthURLWithOpts.
func (mr *MockAuthenticatorMockRecorder) AuthURLWithOpts(state interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{state}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthURLWithOpts", reflect.TypeOf((*MockAuthenticator)(nil).AuthURLWithOpts), varargs...)
}

// Exchange mocks base method.
func (m *MockAuthenticator) Exchange(arg0 string, arg1 ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exchange", varargs...)
	ret0, _ := ret[0].(*oauth2.Token)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exchange indicates an expected call of Exchange.
func (mr *MockAuthenticatorMockRecorder) Exchange(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exchange", reflect.TypeOf((*MockAuthenticator)(nil).Exchange), varargs...)
}

// NewClient mocks base method.
func (m *MockAuthenticator) NewClient(token *oauth2.Token) spotify.Client {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewClient", token)
	ret0, _ := ret[0].(spotify.Client)
	return ret0
}

// NewClient indicates an expected call of NewClient.
func (mr *MockAuthenticatorMockRecorder) NewClient(token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewClient", reflect.TypeOf((*MockAuthenticator)(nil).NewClient), token)
}

// SetAuthInfo mocks base method.
func (m *MockAuthenticator) SetAuthInfo(clientID, secretKey string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetAuthInfo", clientID, secretKey)
}

// SetAuthInfo indicates an expected call of SetAuthInfo.
func (mr *MockAuthenticatorMockRecorder) SetAuthInfo(clientID, secretKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAuthInfo", reflect.TypeOf((*MockAuthenticator)(nil).SetAuthInfo), clientID, secretKey)
}

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

// AddTracksToLibrary mocks base method.
func (m *MockClient) AddTracksToLibrary(ids ...spotify.ID) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range ids {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "AddTracksToLibrary", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddTracksToLibrary indicates an expected call of AddTracksToLibrary.
func (mr *MockClientMockRecorder) AddTracksToLibrary(ids ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTracksToLibrary", reflect.TypeOf((*MockClient)(nil).AddTracksToLibrary), ids...)
}

// CurrentUser mocks base method.
func (m *MockClient) CurrentUser() (*spotify.PrivateUser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CurrentUser")
	ret0, _ := ret[0].(*spotify.PrivateUser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CurrentUser indicates an expected call of CurrentUser.
func (mr *MockClientMockRecorder) CurrentUser() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentUser", reflect.TypeOf((*MockClient)(nil).CurrentUser))
}

// CurrentUsersTracksOpt mocks base method.
func (m *MockClient) CurrentUsersTracksOpt(opt *spotify.Options) (*spotify.SavedTrackPage, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CurrentUsersTracksOpt", opt)
	ret0, _ := ret[0].(*spotify.SavedTrackPage)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CurrentUsersTracksOpt indicates an expected call of CurrentUsersTracksOpt.
func (mr *MockClientMockRecorder) CurrentUsersTracksOpt(opt interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentUsersTracksOpt", reflect.TypeOf((*MockClient)(nil).CurrentUsersTracksOpt), opt)
}

// SearchOpt mocks base method.
func (m *MockClient) SearchOpt(query string, t spotify.SearchType, opt *spotify.Options) (*spotify.SearchResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchOpt", query, t, opt)
	ret0, _ := ret[0].(*spotify.SearchResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchOpt indicates an expected call of SearchOpt.
func (mr *MockClientMockRecorder) SearchOpt(query, t, opt interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchOpt", reflect.TypeOf((*MockClient)(nil).SearchOpt), query, t, opt)
}

// Token mocks base method.
func (m *MockClient) Token() (*oauth2.Token, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Token")
	ret0, _ := ret[0].(*oauth2.Token)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Token indicates an expected call of Token.
func (mr *MockClientMockRecorder) Token() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Token", reflect.TypeOf((*MockClient)(nil).Token))
}
