package net

import (
	"io"
	"net"
	"time"
)

const (
	network = "tcp"
)

type tcpClient struct{}

func (t *tcpClient) DialTimeout(address string, timeout time.Duration) (io.ReadWriteCloser, error) {
	return net.DialTimeout(network, address, timeout)
}

func NewTcpClient() *tcpClient {
	return &tcpClient{}
}
