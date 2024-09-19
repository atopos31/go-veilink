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
}

func NewListener(listenerConfig *config.Listener, sessionMgr *SessionManager, udpSessionMgr *UDPSessionManage) *Listener {
	return &Listener{
		listenerConfig: listenerConfig,
		close:          make(chan struct{}),
		sessionMgr:     sessionMgr,
		udpSessionMgr:  udpSessionMgr,
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

	if l.Encrypt {
		key, err := pkg.GenChacha20Key()
		if err != nil {
			return err
		}
		strKey := pkg.KeyByteToString(key)
		if l.listenerConfig.EncryptKeyPath == "" {
			pkg.WriteKeyToFile(pkg.DefaultKeyPath, l.listenerConfig.ClientID, strKey)
		} else {
			pkg.WriteKeyToFile(l.listenerConfig.EncryptKeyPath, l.listenerConfig.ClientID, strKey)
		}
		l.Key = key
	}
	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			return err
		}
		go l.handleConn(conn)
	}
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
				logrus.Warn(fmt.Sprintf("get session fail: %v", err))
				continue
			}

			// 封装VeilinkProtocol
			pp := &pkg.VeilinkProtocol{
				ClientID:         l.listenerConfig.ClientID,
				PublicProtocol:   l.listenerConfig.PublicProtocol,
				PublicIP:         l.listenerConfig.PublicIP,
				PublicPort:       l.listenerConfig.PublicPort,
				InternalProtocol: l.listenerConfig.InternalProtocol,
				InternalIP:       l.listenerConfig.InternalIP,
				InternalPort:     l.listenerConfig.InternalPort,
			}
			ppBody, err := pp.Encode()
			if err != nil {
				logrus.Warn(fmt.Sprintf("encode proxyprotocol fail: %v", err))
				continue
			}

			tunnelConn.SetWriteDeadline(time.Now().Add(writeTimeout))
			_, err = tunnelConn.Write(ppBody)
			tunnelConn.SetWriteDeadline(time.Time{})
			if err != nil {
				logrus.Warn(fmt.Sprintf("write proxyprotocol fail: %v", err))
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
			logrus.Warn(fmt.Sprintf("encode udp packet fail: %v", err))
			continue
		}
		_, err = udpSess.tunnelConn.Write(body)
		if err != nil {
			logrus.Warn(fmt.Sprintf("write udp packet fail: %v", err))
			continue
		}
	}
	return nil
}

func (l *Listener) udpReadFormClient(tunnelconn net.Conn, raddr net.Addr, conn net.PacketConn) {
	buffer := pkg.UDPpacket{}
	for {
		err := buffer.Decode(tunnelconn)
		if err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				logrus.Warn(fmt.Sprintf("tunnelconn already close: %v", raddr))
				break
			}
			logrus.Warn(fmt.Sprintf("decode udp packet fail: %v", err))
			break
		}

		_, err = conn.WriteTo(buffer, raddr)
		if err != nil {
			logrus.Warn(fmt.Sprintf("write udp packet fail: %v", err))
			break
		}
	}
}

func (l *Listener) handleConn(conn net.Conn) {
	defer conn.Close()

	// 查询session
	tunnelConn, err := l.sessionMgr.GetSessionConnByID(l.listenerConfig.ClientID)
	if err != nil {
		logrus.Warn(fmt.Sprintf("get session fail: %v", err))
		return
	}
	var tunnelConnX pkg.VeilConn
	if l.Encrypt {
		tunnelConnX, err = pkg.NewChacha20Stream(l.Key, tunnelConn)
		if err != nil {
			logrus.Warn(fmt.Sprintf("new chacha20 stream fail: %v", err))
			return
		}
	} else {
		tunnelConnX = tunnelConn
	}
	defer tunnelConnX.Close()

	// 封装VeilinkProtocol
	pp := &pkg.VeilinkProtocol{
		ClientID:         l.listenerConfig.ClientID,
		PublicProtocol:   l.listenerConfig.PublicProtocol,
		PublicIP:         l.listenerConfig.PublicIP,
		PublicPort:       l.listenerConfig.PublicPort,
		InternalProtocol: l.listenerConfig.InternalProtocol,
		InternalIP:       l.listenerConfig.InternalIP,
		InternalPort:     l.listenerConfig.InternalPort,
	}
	ppBody, err := pp.Encode()
	if err != nil {
		logrus.Warn(fmt.Sprintf("encode proxyprotocol fail: %v", err))
		return
	}

	tunnelConnX.SetWriteDeadline(time.Now().Add(writeTimeout))
	_, err = tunnelConn.Write(ppBody)
	tunnelConn.SetWriteDeadline(time.Time{})
	if err != nil {
		logrus.Warn(fmt.Sprintf("write proxyprotocol fail: %v", err))
		return
	}

	in, out := pkg.Join(conn, tunnelConn)
	logrus.Infof("%s in: %d bytes, out: %d bytes", l.listenerConfig.ClientID, in, out)
}

func (l *Listener) Close() {
	l.closeOnce.Do(func() {
		close(l.close)
		if l.listener != nil {
			l.listener.Close()
		}
	})
}
