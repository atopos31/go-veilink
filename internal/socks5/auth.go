package socks5

import (
	"errors"
	"io"
)

const (
	socks5Version = 0x05
)

var (
	UnsupportedProtocol = errors.New("Unsupported socks protocol!")
)

type method = byte

type ClientAuthMessage struct {
	Version byte
	NMethod byte
	Methods []method
}

func NewClientAuthMessage(conn io.Reader) (*ClientAuthMessage, error) {
	// 读取版本
	buf := make([]byte, 2)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		return nil, err
	}

	if buf[0] != socks5Version {
		return nil, UnsupportedProtocol
	}

	nmethods := buf[1]
	methodsBuf := make([]method, nmethods)
	_, err = io.ReadFull(conn, methodsBuf)
	if err != nil {
		return nil, err
	}

	return &ClientAuthMessage{
		Version: socks5Version,
		NMethod: nmethods,
		Methods: methodsBuf,
	}, nil
}
