package config_test

import "github.com/stretchr/testify/mock"

type MockFetcher struct {
	mock.Mock
}

func (m *MockFetcher) Get(key string) (val string) {
	return m.Called(key).String(0)
}
