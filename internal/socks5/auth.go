package socks5

import (
	"errors"
	"io"
)

var (
	UnsupportedProtocol = errors.New("Unsupported socks protocol!")
	UnsupportMethod     = errors.New("Unsupported auth method!")
)

type method = byte

/*
			+----+----------+----------+
	        |VER | NMETHODS | METHODS  |
	        +----+----------+----------+
	        | 1  |    1     | 1 to 255 |
	        +----+----------+----------+
*/
type ClientAuthMessage struct {
	Version byte
	NMethod byte
	Methods []method
}

const (
	NoAuth   method = 0x00
	GSSAPI   method = 0x01
	UserPass method = 0x02
	NoAccept method = 0xff
)

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
