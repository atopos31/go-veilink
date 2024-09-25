package socks5

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

type SOCKS5Server struct {
	IP   string
	Port int
}

func (s *SOCKS5Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.IP, s.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Errorf("socks5 server accept error: %v", err)
			continue
		}

		go func() {
			defer conn.Close()
			if err := handleConnection(conn); err != nil {
				logrus.Errorf("handleConnection error: %v", err)
			}
		}()
	}
}

func handleConnection(conn net.Conn) error {
	// 协商handleConnection
	if err := auth(conn); err != nil {
		return err
	}

	// 请求handleConnection

	// 转发handleConnection

	return nil
}

func auth(conn net.Conn) error {
	_, err := NewClientAuthMessage(conn)
	if err != nil {
		return err
	}
	return nil
}
