package server

import (
	"fmt"
	"net"

	"github.com/atopos31/go-veilink/internal/config"
	"github.com/atopos31/go-veilink/pkg"
	"github.com/sirupsen/logrus"
)

type Gateway struct {
	addr        string
	listenerMgr *ListenerMgr
	sessionMgr  *SessionManager
}

func NewGateway(conf config.Gateway, listenerMgr *ListenerMgr, sessionMgr *SessionManager) *Gateway {
	addr := fmt.Sprintf("%s:%d", conf.Ip, conf.Port)
	return &Gateway{
		addr:        addr,
		listenerMgr: listenerMgr,
		sessionMgr:  sessionMgr,
	}
}

func (g *Gateway) Run() error {
	gateWayListener, err := net.Listen("tcp", g.addr)
	if err != nil {
		return err
	}

	logrus.Debugf("Gateway is running on %s", g.addr)
	go func() {
		defer gateWayListener.Close()
		for {
			conn, err := gateWayListener.Accept()
			if err != nil {
				logrus.Errorf("failed to accept connection %v", err)
				continue
			}
			logrus.Debugf("accept connection from %s", conn.RemoteAddr())
			go g.handleConn(conn)
		}
	}()
	return nil
}

func (g *Gateway) handleConn(conn net.Conn) {
	handshakeReq := &pkg.HandshakeReq{}
	if err := handshakeReq.Decode(conn); err != nil {
		logrus.Errorf("failed to decode handshake request %v", err)
		conn.Close()
		return
	}

	if !g.listenerMgr.CheckExist(handshakeReq.ClientID) {
		logrus.Errorf("invalid client id %v", handshakeReq.ClientID)
		conn.Close()
		return
	}

	logrus.Debugf("handshake request %v", handshakeReq)

	if _, err := g.sessionMgr.AddSession(handshakeReq.ClientID, conn); err != nil {
		logrus.Errorf("failed to add session %v", err)
		return
	}
}

// func (gw *Gateway) DebugInfoTicker(d time.Duration) {
// 	ticker := time.NewTicker(d)
// 	defer ticker.Stop()
// 	for range ticker.C {
// 		gw.sessionMgr.mu.Lock()
// 		logrus.Debugf("↓↓ client is online: %d/%d", len(gw.sessionMgr.sessions), gw.clientCount)
// 		for k := range gw.sessionMgr.sessions {
// 			logrus.Debug(k)
// 		}
// 		gw.sessionMgr.mu.Unlock()
// 	}
// }
