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
	addr       string
	clientIDs  map[string]struct{}
	sessionMgr *SessionManager
}

func NewGateway(conf config.ServerConfig, sessionMgr *SessionManager) *Gateway {
	clientIDsMap := make(map[string]struct{})
	for _, listener := range conf.ListenerConfigs {
		clientIDsMap[listener.ClientID] = struct{}{}
	}
	addr := fmt.Sprintf("%s:%d", conf.Gateway.Ip, conf.Gateway.Port)
	return &Gateway{
		addr:       addr,
		clientIDs:  clientIDsMap,
		sessionMgr: sessionMgr,
	}
}

func (g *Gateway) Run() error {
	go g.checkOnlineInterval()

	gatWayListener, errr := net.Listen("tcp", g.addr)
	if errr != nil {
		return errr
	}
	defer gatWayListener.Close()
	logrus.Debug(fmt.Sprintf("Gateway is running on %s",g.addr))

	for {
		conn, err := gatWayListener.Accept()
		if err != nil {
			return err
		}
		go g.handleConn(conn)
	}
}

func (g *Gateway) handleConn(conn net.Conn) {
	handshakeReq := &pkg.HandshakeReq{}
	if err := handshakeReq.Decode(conn); err != nil {
		logrus.Error(fmt.Sprintf("failed to decode handshake request %v", err))
		conn.Close()
		return
	}
	if _, ok := g.clientIDs[handshakeReq.ClientID]; !ok {
		logrus.Error(fmt.Sprintf("invalid client id %v", handshakeReq.ClientID))
		conn.Close()
		return
	}

	logrus.Debug(fmt.Sprintf("handshake request %v", handshakeReq))

	if _, err := g.sessionMgr.AddSession(handshakeReq.ClientID, conn); err != nil {
		logrus.Error(fmt.Sprintf("failed to add session %v", err))
		return
	}
}

func (gw *Gateway) checkOnlineInterval() {
	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()
	for range tick.C {
		logrus.Debug(fmt.Sprintf("online client count %d/%d",len(gw.sessionMgr.sessions),gw.sessionMgr.count))
		gw.sessionMgr.Range(func(k string, v *Session) bool {
			if v.Connection.IsClosed() {
				logrus.Debug(fmt.Sprintf("session %s is offline", v.ClientID))
				return false
			}

			logrus.Debug(fmt.Sprintf("session %s is online", v.ClientID))
			return true
		})
	}
}
