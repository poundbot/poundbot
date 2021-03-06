// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import storage "github.com/poundbot/poundbot/storage"
import types "github.com/poundbot/poundbot/types"

// DiscordAuthsStore is an autogenerated mock type for the DiscordAuthsStore type
type DiscordAuthsStore struct {
	mock.Mock
}

// GetByDiscordID provides a mock function with given fields: snowflake
func (_m *DiscordAuthsStore) GetByDiscordID(snowflake string) (types.DiscordAuth, error) {
	ret := _m.Called(snowflake)

	var r0 types.DiscordAuth
	if rf, ok := ret.Get(0).(func(string) types.DiscordAuth); ok {
		r0 = rf(snowflake)
	} else {
		r0 = ret.Get(0).(types.DiscordAuth)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(snowflake)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByDiscordName provides a mock function with given fields: discordName
func (_m *DiscordAuthsStore) GetByDiscordName(discordName string) (types.DiscordAuth, error) {
	ret := _m.Called(discordName)

	var r0 types.DiscordAuth
	if rf, ok := ret.Get(0).(func(string) types.DiscordAuth); ok {
		r0 = rf(discordName)
	} else {
		r0 = ret.Get(0).(types.DiscordAuth)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(discordName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Remove provides a mock function with given fields: _a0
func (_m *DiscordAuthsStore) Remove(_a0 storage.UserInfoGetter) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(storage.UserInfoGetter) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Upsert provides a mock function with given fields: _a0
func (_m *DiscordAuthsStore) Upsert(_a0 types.DiscordAuth) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(types.DiscordAuth) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
