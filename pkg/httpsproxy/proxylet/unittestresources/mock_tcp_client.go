package unittestresources

import (
	"io"
	"sync"
	"time"
)

type byteReadWriteCloser struct {
	lock          sync.Mutex
	readerPointer int
	ReadQueue     []string
	WriteQueue    []string
	closed        bool
}

func (b *byteReadWriteCloser) waitForData() {
	for {
		if b.closed || b.readerPointer < len(b.ReadQueue) {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (b *byteReadWriteCloser) Read(p []byte) (n int, err error) {
	b.waitForData()

	b.lock.Lock()
	defer b.lock.Unlock()

	if b.closed {
		return 0, io.EOF
	}

	data := b.ReadQueue[b.readerPointer]
	b.readerPointer++
	p = []byte(data)
	return len(data), nil
}

func (b *byteReadWriteCloser) Write(p []byte) (n int, err error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.WriteQueue = append(b.WriteQueue, string(p))
	return len(p), nil
}

func (b *byteReadWriteCloser) Close() error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.closed = true
	return nil
}

type mockTcpClient struct {
	Buffer *byteReadWriteCloser
}

func (m *mockTcpClient) DialTimeout(address string, timeout time.Duration) (io.ReadWriteCloser, error) {
	return m.Buffer, nil
}

func NewMockTcpClient() *mockTcpClient {
	return &mockTcpClient{
		Buffer: &byteReadWriteCloser{},
	}
}
