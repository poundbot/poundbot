// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import mock "github.com/stretchr/testify/mock"

import types "bitbucket.org/mrpoundsign/poundbot/types"

// UsersStore is an autogenerated mock type for the UsersStore type
type UsersStore struct {
	mock.Mock
}

// Get provides a mock function with given fields: steamID, u
func (_m *UsersStore) Get(steamID uint64, u *types.User) error {
	ret := _m.Called(steamID, u)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint64, *types.User) error); ok {
		r0 = rf(steamID, u)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveClan provides a mock function with given fields: serverKey, tag
func (_m *UsersStore) RemoveClan(serverKey string, tag string) error {
	ret := _m.Called(serverKey, tag)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(serverKey, tag)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveClansNotIn provides a mock function with given fields: serverKey, tags
func (_m *UsersStore) RemoveClansNotIn(serverKey string, tags []string) error {
	ret := _m.Called(serverKey, tags)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []string) error); ok {
		r0 = rf(serverKey, tags)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetClanIn provides a mock function with given fields: serverKey, tag, steamIds
func (_m *UsersStore) SetClanIn(serverKey string, tag string, steamIds []uint64) error {
	ret := _m.Called(serverKey, tag, steamIds)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, []uint64) error); ok {
		r0 = rf(serverKey, tag, steamIds)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpsertBase provides a mock function with given fields: baseUser
func (_m *UsersStore) UpsertBase(baseUser types.BaseUser) error {
	ret := _m.Called(baseUser)

	var r0 error
	if rf, ok := ret.Get(0).(func(types.BaseUser) error); ok {
		r0 = rf(baseUser)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}