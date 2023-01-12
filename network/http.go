package network

import (
	"io"
	"net/http"
)

type Http interface {
	Get(req *http.Request) ([]byte, error)
}

type DefaultHttp struct {
	client *http.Client
}

func NewHttp() Http {
	return &DefaultHttp{
		client: &http.Client{},
	}
}

func (d *DefaultHttp) Get(req *http.Request) ([]byte, error) {
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)

	return buf, err
}
