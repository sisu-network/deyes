package network

import "net/http"

type MockHttp struct {
	GetFunc func(req *http.Request) ([]byte, error)
}

func (m *MockHttp) Get(req *http.Request) ([]byte, error) {
	if m.GetFunc != nil {
		return m.GetFunc(req)
	}

	return nil, nil
}
