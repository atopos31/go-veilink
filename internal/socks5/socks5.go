package socks5

import (
	"fmt"
	"net"
	"slices"

	"github.com/atopos31/go-veilink/pkg"
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

	logrus.Infof("socks5 server listen on %s", addr)

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
	req, err := s.request(conn)
	if err != nil {
		return err
	}

	replies := RepliesTODO()
	remoteConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", req.DstAddr, req.DstPort))
	if err != nil {
		logrus.Debugf("remote dial addr%s port%d error: %v", req.DstAddr, req.DstPort, err)
		if err := replies.WithREP(REP_HOST_UNREACHABLE).WithATYP(req.Atyp).Wirte(conn); err != nil {
			return err
		}
		return err
	}
	defer remoteConn.Close()

	logrus.Debugf("localaddr: %s", remoteConn.LocalAddr())

	if err := replies.WithREP(REP_SUCCEEDED).WithATYP(req.Atyp).Wirte(conn); err != nil {
		return err
	}

	// 转发handleConnection
	in, out := pkg.Join(conn, remoteConn)
	logrus.Debugf("in:%d out:%d", in, out)
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

func (s *SOCKS5Server) request(conn net.Conn) (*ClientRequestMessage, error) {
	replies := RepliesTODO()
	req, err := NewClientRequestMessage(conn)
	if err != nil {
		if err := replies.WithREPByError(err).Wirte(conn); err != nil {
			return nil, err
		}
		return nil, err
	}

	return req, nil
}
