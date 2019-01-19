// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import storage "bitbucket.org/mrpoundsign/poundbot/storage"

// Storage is an autogenerated mock type for the Storage type
type Storage struct {
	mock.Mock
}

// Accounts provides a mock function with given fields:
func (_m *Storage) Accounts() storage.AccountsStore {
	ret := _m.Called()

	var r0 storage.AccountsStore
	if rf, ok := ret.Get(0).(func() storage.AccountsStore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storage.AccountsStore)
		}
	}

	return r0
}

// Close provides a mock function with given fields:
func (_m *Storage) Close() {
	_m.Called()
}

// Copy provides a mock function with given fields:
func (_m *Storage) Copy() storage.Storage {
	ret := _m.Called()

	var r0 storage.Storage
	if rf, ok := ret.Get(0).(func() storage.Storage); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storage.Storage)
		}
	}

	return r0
}

// DiscordAuths provides a mock function with given fields:
func (_m *Storage) DiscordAuths() storage.DiscordAuthsStore {
	ret := _m.Called()

	var r0 storage.DiscordAuthsStore
	if rf, ok := ret.Get(0).(func() storage.DiscordAuthsStore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storage.DiscordAuthsStore)
		}
	}

	return r0
}

// Init provides a mock function with given fields:
func (_m *Storage) Init() {
	_m.Called()
}

// RaidAlerts provides a mock function with given fields:
func (_m *Storage) RaidAlerts() storage.RaidAlertsStore {
	ret := _m.Called()

	var r0 storage.RaidAlertsStore
	if rf, ok := ret.Get(0).(func() storage.RaidAlertsStore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storage.RaidAlertsStore)
		}
	}

	return r0
}

// Users provides a mock function with given fields:
func (_m *Storage) Users() storage.UsersStore {
	ret := _m.Called()

	var r0 storage.UsersStore
	if rf, ok := ret.Get(0).(func() storage.UsersStore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storage.UsersStore)
		}
	}

	return r0
}
