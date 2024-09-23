package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/atopos31/go-veilink/internal/config"
	"github.com/atopos31/go-veilink/pkg"
	"github.com/sirupsen/logrus"
)

var (
	writeTimeout = time.Second * 3
)

type Listener struct {
	listenerConfig *config.Listener
	Encrypt        bool
	Key            []byte
	sessionMgr     *SessionManager
	closeOnce      sync.Once
	close          chan struct{}
	listener       net.Listener
	udpSessionMgr  *UDPSessionManage
	ioData         *IOdata
}

func NewListener(listenerConfig *config.Listener, keymap *keymap, sessionMgr *SessionManager, udpSessionMgr *UDPSessionManage) *Listener {
	var key []byte
	var err error
	if listenerConfig.Encrypt {
		key, err = keymap.Get(listenerConfig.ClientID)
		if err != nil {
			panic(err)
		}
	}
	return &Listener{
		Encrypt:        listenerConfig.Encrypt,
		Key:            key,
		listenerConfig: listenerConfig,
		close:          make(chan struct{}),
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
	defer tcpListener.Close()

	l.listener = tcpListener

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			return err
		}
		go l.handleConn(conn)
	}
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
		tunnelConn, err = pkg.NewChacha20Stream(l.Key, tunnelConn)
		if err != nil {
			logrus.Warnf("new chacha20 stream fail: %v", err)
			return
		}
	}
	if err := l.sendVeilinkProtocol(tunnelConn); err != nil {
		logrus.Warnf("send veilink protocol fail: %v", err)
		return
	}

	in, out := pkg.Join(conn, tunnelConn)
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
	defer udpListener.Close()
	buffer := make([]byte, 1024*64)
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
				tunnelConn, err = pkg.NewChacha20Stream(l.Key, tunnelConn)
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

		packet := pkg.UDPpacket(buffer[:n])
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
	return nil
}

func (l *Listener) udpReadFormClient(tunnelconn pkg.VeilConn, raddr net.Addr, conn net.PacketConn) {
	buffer := pkg.UDPpacket{}
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
func (l *Listener) sendEncryptProtocol(conn pkg.VeilConn) error {
	enc := &pkg.EncryptProtocl{}
	encByte := enc.Encode(l.Encrypt)
	conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	_, err := conn.Write(encByte)
	conn.SetWriteDeadline(time.Time{})
	return err
}

// Inform the client of the specific IP and port that need to be tunneled into the internal network.
func (l *Listener) sendVeilinkProtocol(conn pkg.VeilConn) error {
	pp := &pkg.VeilinkProtocol{
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
		close(l.close)
		if l.listener != nil {
			l.listener.Close()
		}
	})
}

func (l *Listener) PrintDebugInfo() {
	logrus.Debug(fmt.Sprintf("Listener: server %s:%d <=Veilink %s=>client %s %s:%d",
		l.listenerConfig.PublicIP,
		l.listenerConfig.PublicPort,
		l.listenerConfig.PublicProtocol,
		l.listenerConfig.ClientID,
		l.listenerConfig.InternalIP,
		l.listenerConfig.InternalPort,
	))
}
