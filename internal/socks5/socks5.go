package socks5

import (
	"fmt"
	"net"
	"slices"

	"github.com/sirupsen/logrus"
)

const (
	socks5Version = 0x05
)

type SOCKS5Server struct {
	IP         string
	Port       int
	AuthMethod method
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
		logrus.Debugf("new connection from %s", conn.RemoteAddr())

		go func() {
			defer conn.Close()
			if err := s.handleConnection(conn); err != nil {
				logrus.Errorf("handleConnection error: %v", err)
			}
		}()
	}
}

func (s *SOCKS5Server) handleConnection(conn net.Conn) error {
	// 协商handleConnection
	if err := s.auth(conn); err != nil {
		return err
	}

	// 请求handleConnection

	// 转发handleConnection

	return nil
}

func (s *SOCKS5Server) auth(conn net.Conn) error {
	clientMsg, err := NewClientAuthMessage(conn)
	if err != nil {
		return err
	}
	logrus.Debugf("clientMsg: %v", clientMsg)

	serverMsg := []byte{socks5Version, s.AuthMethod}

	// 如果服务端没有支持的认证方式
	if !slices.Contains(clientMsg.Methods, s.AuthMethod) {
		serverMsg[1] = NoAccept
		if _, err = conn.Write(serverMsg); err != nil {
			return err
		}
		return UnsupportMethod
	}

	_, err = conn.Write(serverMsg)
	return err
}
