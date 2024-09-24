package server

import (
	"fmt"
	"net"
	"time"

	"github.com/atopos31/go-veilink/internal/config"
	"github.com/atopos31/go-veilink/pkg"
	"github.com/sirupsen/logrus"
)

type Gateway struct {
	addr        string
	clientCount int
	clientIDs   map[string]struct{}
	sessionMgr  *SessionManager
}

func NewGateway(conf config.ServerConfig, sessionMgr *SessionManager) *Gateway {
	clientIDsMap := make(map[string]struct{})
	for _, listener := range conf.ListenerConfigs {
		clientIDsMap[listener.ClientID] = struct{}{}
	}
	addr := fmt.Sprintf("%s:%d", conf.Gateway.Ip, conf.Gateway.Port)
	return &Gateway{
		addr:        addr,
		clientCount: len(clientIDsMap),
		clientIDs:   clientIDsMap,
		sessionMgr:  sessionMgr,
	}
}

func (g *Gateway) Run() error {
	gateWayListener, errr := net.Listen("tcp", g.addr)
	if errr != nil {
		return errr
	}
	defer gateWayListener.Close()
	logrus.Debugf("Gateway is running on %s", g.addr)

	for {
		conn, err := gateWayListener.Accept()
		if err != nil {
			return err
		}
		logrus.Debugf(fmt.Sprintf("accept connection from %s", conn.RemoteAddr()))
		go g.handleConn(conn)
	}
}

func (g *Gateway) handleConn(conn net.Conn) {
	handshakeReq := &pkg.HandshakeReq{}
	if err := handshakeReq.Decode(conn); err != nil {
		logrus.Errorf("failed to decode handshake request %v", err)
		conn.Close()
		return
	}
	if _, ok := g.clientIDs[handshakeReq.ClientID]; !ok {
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

func (gw *Gateway) DebugInfoTicker(d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for range ticker.C {
		gw.sessionMgr.mu.Lock()
		logrus.Debugf("↓↓ client is online: %d/%d", len(gw.sessionMgr.sessions), gw.clientCount)
		for k := range gw.sessionMgr.sessions {
			logrus.Debug(k)
		}
		gw.sessionMgr.mu.Unlock()
	}

}
