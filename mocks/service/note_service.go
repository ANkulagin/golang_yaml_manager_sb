// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	entity "github.com/ANkulagin/golang_yaml_manager_sb/internal/domain/entity"
	mock "github.com/stretchr/testify/mock"
)

// NoteService is an autogenerated mock type for the NoteService type
type NoteService struct {
	mock.Mock
}

// ValidateAndUpdate provides a mock function with given fields: note
func (_m *NoteService) ValidateAndUpdate(note *entity.Note) (bool, error) {
	ret := _m.Called(note)

	if len(ret) == 0 {
		panic("no return value specified for ValidateAndUpdate")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(*entity.Note) (bool, error)); ok {
		return rf(note)
	}
	if rf, ok := ret.Get(0).(func(*entity.Note) bool); ok {
		r0 = rf(note)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(*entity.Note) error); ok {
		r1 = rf(note)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewNoteService creates a new instance of NoteService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNoteService(t interface {
	mock.TestingT
	Cleanup(func())
}) *NoteService {
	mock := &NoteService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
