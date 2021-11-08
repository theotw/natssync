package unittestresources

import (
	"bytes"
	"io"
	"net/http"
)

type mockHttpClient struct {
	httpRequest *http.Request
	httpResponse *http.Response
	httpError error
}

func (m mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	m.httpRequest = req
	return m.httpResponse, m.httpError
}

func NewMockHttpClient() *mockHttpClient {

	return &mockHttpClient{
		httpResponse: &http.Response{
			StatusCode:       200,
			Body:             io.NopCloser(bytes.NewReader([]byte("ok"))),
		},
	}
}
