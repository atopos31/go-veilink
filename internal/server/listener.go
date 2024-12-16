package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/atopos31/go-veilink/internal/common"
	"github.com/atopos31/go-veilink/internal/config"
	"github.com/sirupsen/logrus"
)

var (
	writeTimeout = time.Second * 3
)

type Listener struct {
	Uuid           string
	listenerConfig *config.Listener
	Encrypt        bool
	Key            []byte
	sessionMgr     *SessionManager
	closeOnce      sync.Once
	listener       common.ConnWithClose
	udpSessionMgr  *UDPSessionManage
	ioData         *IOdata
}

func NewListener(listenerConfig *config.Listener, key []byte, sessionMgr *SessionManager, udpSessionMgr *UDPSessionManage) *Listener {
	return &Listener{
		Uuid:           listenerConfig.Uuid,
		Encrypt:        listenerConfig.Encrypt,
		Key:            key,
		listenerConfig: listenerConfig,
		sessionMgr:     sessionMgr,
		udpSessionMgr:  udpSessionMgr,
		ioData:         new(IOdata),
	}
}

func (l *Listener) ListenAndServe() error {
	switch l.listenerConfig.PublicProtocol {
	case "tcp":
		return l.listenerAndServerTCP()
	case "udp":
		return l.listenerAndServerUDP()
	default:
		return fmt.Errorf("TODO://")
	}
}

func (l *Listener) listenerAndServerTCP() error {
	addr := fmt.Sprintf("%s:%d", l.listenerConfig.PublicIP, l.listenerConfig.PublicPort)
	tcpListener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	l.listener = tcpListener

	go func() {
		defer tcpListener.Close()
		for {
			conn, err := tcpListener.Accept()
			if err != nil {
				return
			}
			go l.handleConn(conn)
		}
	}()
	return nil
}

func (l *Listener) handleConn(conn net.Conn) {
	defer conn.Close()

	// search session
	tunnelConn, err := l.sessionMgr.GetSessionConnByID(l.listenerConfig.ClientID)
	if err != nil {
		logrus.Warnf("get session fail: %v", err)
		return
	}
	if err := l.sendEncryptProtocol(tunnelConn); err != nil {
		logrus.Warnf("send encrypt protocol fail: %v", err)
		return
	}
	if l.Encrypt {
		tunnelConn, err = common.NewChacha20Stream(l.Key, tunnelConn)
		if err != nil {
			logrus.Warnf("new chacha20 stream fail: %v", err)
			return
		}
	}
	if err := l.sendVeilinkProtocol(tunnelConn); err != nil {
		logrus.Warnf("send veilink protocol fail: %v", err)
		return
	}

	in, out := common.Join(conn, tunnelConn)
	// add io data
	l.ioData.AddInput(in)
	l.ioData.AddOutput(out)
	logrus.Infof("%s in: %d bytes, out: %d bytes", l.listenerConfig.ClientID, in, out)
}

func (l *Listener) listenerAndServerUDP() error {
	addr := fmt.Sprintf("%s:%d", l.listenerConfig.PublicIP, l.listenerConfig.PublicPort)
	udpListener, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}

	buffer := make([]byte, 1024*64)
	l.listener = udpListener
	go func() {
		defer udpListener.Close()
		for {
			n, remoteAddr, err := udpListener.ReadFrom(buffer)
			if err != nil {
				break
			}
			udpSess, err := l.udpSessionMgr.Get(remoteAddr.String())
			if err != nil {
				// 查询session
				tunnelConn, err := l.sessionMgr.GetSessionConnByID(l.listenerConfig.ClientID)
				if err != nil {
					logrus.Warnf("get session fail: %v", err)
					continue
				}

				if err := l.sendEncryptProtocol(tunnelConn); err != nil {
					logrus.Warnf("send encrypt protocol fail: %v", err)
					continue
				}

				if l.Encrypt {
					tunnelConn, err = common.NewChacha20Stream(l.Key, tunnelConn)
					if err != nil {
						logrus.Warnf("new chacha20 stream fail: %v", err)
						continue
					}
				}

				if err := l.sendVeilinkProtocol(tunnelConn); err != nil {
					logrus.Warnf("send encrypt protocol fail: %v", err)
					continue
				}

				udpSess = &UDPsession{
					tunnelConn: tunnelConn,
					LocalAddr:  addr,
					RemoteAddr: remoteAddr.String(),
				}
				l.udpSessionMgr.Add(remoteAddr.String(), udpSess)
				go l.udpReadFormClient(tunnelConn, remoteAddr, udpListener)
			}

			packet := common.UDPpacket(buffer[:n])
			body, err := packet.Encode()
			if err != nil {
				logrus.Warnf("encode udp packet fail: %v", err)
				continue
			}
			lenbody, err := udpSess.tunnelConn.Write(body)
			if err != nil {
				logrus.Warnf("write udp packet fail: %v", err)
				continue
			}
			l.ioData.AddInput(int64(lenbody))
		}
	}()

	return nil
}

func (l *Listener) udpReadFormClient(tunnelconn common.VeilConn, raddr net.Addr, conn net.PacketConn) {
	buffer := common.UDPpacket{}
	for {
		err := buffer.Decode(tunnelconn)
		if err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				logrus.Warnf("tunnelconn already close: %v", raddr)
				break
			}
			logrus.Warnf("decode udp packet fail: %v", err)
			break
		}

		lenbuffer, err := conn.WriteTo(buffer, raddr)
		if err != nil {
			logrus.Warnf("write udp packet fail: %v", err)
			break
		}
		l.ioData.AddOutput(int64(lenbuffer))
	}
}

// Inform the client whether the current connection is encrypted.
func (l *Listener) sendEncryptProtocol(conn common.VeilConn) error {
	enc := &common.EncryptProtocl{}
	encByte := enc.Encode(l.Encrypt)
	conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	_, err := conn.Write(encByte)
	conn.SetWriteDeadline(time.Time{})
	return err
}

// Inform the client of the specific IP and port that need to be tunneled into the internal network.
func (l *Listener) sendVeilinkProtocol(conn common.VeilConn) error {
	pp := &common.VeilinkProtocol{
		ClientID:       l.listenerConfig.ClientID,
		PublicProtocol: l.listenerConfig.PublicProtocol,
		PublicIP:       l.listenerConfig.PublicIP,
		PublicPort:     l.listenerConfig.PublicPort,
		InternalIP:     l.listenerConfig.InternalIP,
		InternalPort:   l.listenerConfig.InternalPort,
	}
	ppBody, err := pp.Encode()
	if err != nil {
		return err
	}
	conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	_, err = conn.Write(ppBody)
	conn.SetWriteDeadline(time.Time{})
	return err
}
func (l *Listener) Close() {
	l.closeOnce.Do(func() {
		if l.listener != nil {
			l.listener.Close()
		}
	})
}
