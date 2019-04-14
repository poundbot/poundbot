// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

import time "time"
import types "github.com/poundbot/poundbot/types"

// RaidAlertsStore is an autogenerated mock type for the RaidAlertsStore type
type RaidAlertsStore struct {
	mock.Mock
}

// AddInfo provides a mock function with given fields: alertIn, ed
func (_m *RaidAlertsStore) AddInfo(alertIn time.Duration, ed types.EntityDeath) error {
	ret := _m.Called(alertIn, ed)

	var r0 error
	if rf, ok := ret.Get(0).(func(time.Duration, types.EntityDeath) error); ok {
		r0 = rf(alertIn, ed)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetReady provides a mock function with given fields:
func (_m *RaidAlertsStore) GetReady() ([]types.RaidAlert, error) {
	ret := _m.Called()

	var r0 []types.RaidAlert
	if rf, ok := ret.Get(0).(func() []types.RaidAlert); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]types.RaidAlert)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Remove provides a mock function with given fields: _a0
func (_m *RaidAlertsStore) Remove(_a0 types.RaidAlert) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(types.RaidAlert) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
