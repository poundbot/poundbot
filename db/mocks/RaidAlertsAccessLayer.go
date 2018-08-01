// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import mock "github.com/stretchr/testify/mock"
import types "mrpoundsign.com/poundbot/types"

// RaidAlertsAccessLayer is an autogenerated mock type for the RaidAlertsAccessLayer type
type RaidAlertsAccessLayer struct {
	mock.Mock
}

// AddInfo provides a mock function with given fields: _a0
func (_m *RaidAlertsAccessLayer) AddInfo(_a0 types.EntityDeath) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(types.EntityDeath) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetReady provides a mock function with given fields: _a0
func (_m *RaidAlertsAccessLayer) GetReady(_a0 *[]types.RaidNotification) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*[]types.RaidNotification) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Remove provides a mock function with given fields: _a0
func (_m *RaidAlertsAccessLayer) Remove(_a0 types.RaidNotification) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(types.RaidNotification) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}