package utils

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type responseWriter struct {
	gin.ResponseWriter
	body   []byte
	header http.Header
	status int
}

func NewResponseWriter(ginContextResponseWriter gin.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: ginContextResponseWriter,
		header:         http.Header{},
	}
}

func (rw *responseWriter) Write(body []byte) (int, error) {
	rw.body = append(rw.body, body...)
	return len(rw.body), nil
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
}

func (rw *responseWriter) WriteString(data string) (int, error) {
	rw.body = append(rw.body, []byte(data)...)
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

func (rw *responseWriter) Status() int {
	return rw.GetStatus()
}

func (rw *responseWriter) Size() int {
	return len(rw.body)
}

func (rw *responseWriter) WriteHeaderNow() {}

func (rw *responseWriter) WriteDataOut() {

	// add back the headers
	for key, values := range rw.header {
		for _, value := range values {
			rw.ResponseWriter.Header().Add(key, value)
		}
	}

	// write the headers
	rw.ResponseWriter.WriteHeader(rw.status)

	// write the body
	if _, err := rw.ResponseWriter.Write(rw.body); err != nil {
		log.WithError(err).Errorf("failed to write body to gin response writer")
	}
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	fmt.Printf("responseWriter.body: \"%s\"\n", string(rw.body[:]))
	rw.WriteDataOut()
	return rw.ResponseWriter.Hijack()
}
