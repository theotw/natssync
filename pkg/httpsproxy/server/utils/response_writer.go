package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type responseWriter struct {
	gin.ResponseWriter
	body   []byte
	header http.Header
	status int
}

func NewResponseWriter() *responseWriter {
	return &responseWriter{
		header: http.Header{},
	}
}

func (rw *responseWriter) Write(body []byte) (int, error) {
	rw.body = body
	return len(rw.body), nil
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
}

func (rw *responseWriter) WriteString(data string) (int, error) {
	rw.body = []byte(data)
	return len(data), nil
}

func (rw *responseWriter) GetStatus() int {
	return rw.status
}

func (rw *responseWriter) GetBody() []byte {
	return rw.body
}

func (rw *responseWriter) GetHeaders() http.Header {
	return rw.header
}
