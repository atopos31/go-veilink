package server

import (
	"fmt"
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
	sessionMgr     *SessionManager
	closeOnce      sync.Once
	close          chan struct{}
	listener       net.Listener
}

func NewListener(listenerConfig *config.Listener, sessionMgr *SessionManager) *Listener {
	return &Listener{
		listenerConfig: listenerConfig,
		close:          make(chan struct{}),
		sessionMgr:     sessionMgr,
	}
}

func (l *Listener) ListenAndServe() error {
	switch l.listenerConfig.PublicProtocol {
	case "tcp":
		return l.listenerAndServerTCP()
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

	// 查询session
	tunnelConn, err := l.sessionMgr.GetSessionConnByID(l.listenerConfig.ClientID)
	if err != nil {
		logrus.Warn(fmt.Sprintf("get session fail: %v", err))
		return
	}
	defer tunnelConn.Close()

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

	tunnelConn.SetWriteDeadline(time.Now().Add(writeTimeout))
	_, err = tunnelConn.Write(ppBody)
	tunnelConn.SetWriteDeadline(time.Time{})
	if err != nil {
		logrus.Warn(fmt.Sprintf("write proxyprotocol fail: %v", err))
		return
	}

	// 双向数据拷贝
	// go func() {
	// 	defer tunnelConn.Close()
	// 	defer conn.Close()
	// 	io.Copy(tunnelConn, conn)
	// }()
	// io.Copy(conn, tunnelConn)
	in,out  := pkg.Join(conn, tunnelConn)
	logrus.Infof("in: %d bytes, out: %d bytes", in, out)
}

func (l *Listener) Close() {
	l.closeOnce.Do(func() {
		close(l.close)
		if l.listener != nil {
			l.listener.Close()
		}
	})
}
